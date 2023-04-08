package listener

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	agent "github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener"
	"github.com/alexandria-oss/streams/proxy/egress"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/viper"
)

var (
	ErrWALUnknownRelation = errors.New("unknown relation ID")
)

const WALDriver = "postgres_wal"

var defaultLoggerWAL = agent.DefaultLogger.With().
	Str("listener_driver", WALDriver).
	Logger()

type walConfig struct {
	OutputPlugin          string
	ConnectionString      string
	PublicationName       string
	SlotName              string
	EgressTable           string
	EnablePublishCreation bool
	LogPollingInterval    time.Duration
	WorkerTimeout         time.Duration
}

func newWALConfig() walConfig {
	viper.SetDefault("postgres.wal.output_plugin", "pgoutput")
	viper.SetDefault("postgres.wal.publication_name", "streams_egress_proxy")
	viper.SetDefault("postgres.wal.slot_name", "streams_egress_proxy_worker")
	viper.SetDefault("postgres.egress_table", "streams_egress")
	viper.SetDefault("postgres.wal.enable_publish_creation", false)
	viper.SetDefault("postgres.wal.log_polling_interval", time.Second*15)
	viper.SetDefault("postgres.wal.worker_timeout", time.Second*5)
	return walConfig{
		OutputPlugin:          viper.GetString("postgres.wal.output_plugin"),
		ConnectionString:      fmt.Sprintf("%s&replication=database", viper.GetString("postgres.connection_string")),
		PublicationName:       viper.GetString("postgres.wal.publication_name"),
		SlotName:              viper.GetString("postgres.wal.slot_name"),
		EgressTable:           viper.GetString("postgres.egress_table"),
		EnablePublishCreation: viper.GetBool("postgres.wal.enable_publish_creation"),
		LogPollingInterval:    viper.GetDuration("postgres.wal.log_polling_interval"),
		WorkerTimeout:         viper.GetDuration("postgres.wal.worker_timeout"),
	}
}

type WAL struct {
	cfg               walConfig
	fwd               egress.Forwarder
	baseCtx           context.Context
	baseCtxCancel     context.CancelFunc
	inFlightProcesses sync.WaitGroup
	totalReads        atomic.Uint64

	conn                  *pgconn.PgConn
	slotConfirmedFlushLSN pglogrepl.LSN
	sysInfo               pglogrepl.IdentifySystemResult

	walReaderOffset pglogrepl.LSN
	commitMsgTime   time.Time
	relationMsgMap  map[uint32]*pglogrepl.RelationMessage
	msgTypeMap      *pgtype.Map
}

var _ Listener = &WAL{}

func NewWAL(fwd egress.Forwarder) *WAL {
	return &WAL{
		cfg:               newWALConfig(),
		fwd:               fwd,
		inFlightProcesses: sync.WaitGroup{},
		totalReads:        atomic.Uint64{},
	}
}

func (w *WAL) Start() error {
	w.baseCtx, w.baseCtxCancel = context.WithCancel(context.Background())
	var err error
	w.conn, err = pgconn.Connect(w.baseCtx, w.cfg.ConnectionString)
	if err != nil {
		return err
	}

	if err = w.createPublication(); err != nil {
		return err
	}

	if err = w.createReplicationSlot(); err != nil {
		return err
	}

	if err = w.fetchSysInfo(); err != nil {
		return err
	}

	if err = w.fetchSlotInfo(); err != nil {
		return err
	}

	return w.startLogListener()
}

func (w *WAL) createPublication() error {
	if !w.cfg.EnablePublishCreation {
		return nil
	}
	query := fmt.Sprintf("CREATE PUBLICATION %s FOR TABLE %s;", w.cfg.PublicationName, w.cfg.EgressTable)
	result := w.conn.Exec(w.baseCtx, query)
	_, err := result.ReadAll()
	if err != nil {
		return err
	}
	defaultLoggerWAL.Info().
		Str("publication_name", w.cfg.PublicationName).
		Str("table_name", w.cfg.EgressTable).
		Msg("created publication for specified table")
	return nil
}

func (w *WAL) createReplicationSlot() error {
	res, errCreate := pglogrepl.CreateReplicationSlot(w.baseCtx, w.conn, w.cfg.SlotName, w.cfg.OutputPlugin, pglogrepl.CreateReplicationSlotOptions{
		Temporary:      false,
		SnapshotAction: "",
		Mode:           pglogrepl.LogicalReplication,
	})

	if errCreate != nil {
		errStr := errCreate.Error()
		isAlreadyExistsErr := strings.HasPrefix(errStr, "ERROR: replication slot") && strings.HasSuffix(errStr, "already exists (SQLSTATE 42710)")
		if !isAlreadyExistsErr {
			return errCreate
		}
		defaultLoggerWAL.Warn().
			Str("slot_name", w.cfg.SlotName).
			Str("output_plugin", w.cfg.OutputPlugin).
			Msg("replication slot already created")
		return nil
	}
	defaultLoggerWAL.Info().
		Str("slot_name", res.SlotName).
		Str("snapshot_name", res.SnapshotName).
		Str("output_plugin", res.OutputPlugin).
		Str("consistent_point", res.ConsistentPoint).
		Msg("created replication slot")
	return nil
}

func (w *WAL) fetchSysInfo() error {
	systemIdentification, err := pglogrepl.IdentifySystem(w.baseCtx, w.conn)
	if err != nil {
		return err
	}
	w.sysInfo = systemIdentification
	defaultLoggerWAL.Info().
		Str("system_id", systemIdentification.SystemID).
		Str("database_name", systemIdentification.DBName).
		Int32("timeline", systemIdentification.Timeline).
		Str("xlogpos", systemIdentification.XLogPos.String()).
		Msg("detected database system")
	return nil
}

func (w *WAL) fetchSlotInfo() error {
	query := fmt.Sprintf("SELECT confirmed_flush_lsn FROM pg_catalog.pg_replication_slots WHERE slot_name = '%s';", w.cfg.SlotName)
	res := w.conn.Exec(w.baseCtx, query)
	if res == nil {
		return nil
	}

	defer func() {
		errClose := res.Close()
		if errClose != nil {
			defaultLoggerWAL.Err(errClose).
				Str("process", "fetchSlotInfo").
				Msg("failed to close result reader")
		}
	}()

	res.NextResult()
	reader := res.ResultReader().Read()
	if reader.Err != nil {
		return reader.Err
	}

	var err error
	w.slotConfirmedFlushLSN, err = pglogrepl.ParseLSN(string(reader.Rows[0][0]))
	if err != nil {
		return err
	}
	defaultLoggerWAL.Info().
		Str("slot_name", w.cfg.SlotName).
		Str("last_read_offset", w.slotConfirmedFlushLSN.String()).
		Msg("fetched last read offset of the slot")
	return nil
}

func (w *WAL) startLogListener() error {
	err := pglogrepl.StartReplication(w.baseCtx, w.conn, w.cfg.SlotName, w.slotConfirmedFlushLSN, pglogrepl.StartReplicationOptions{
		Timeline:   0,
		Mode:       0,
		PluginArgs: []string{"proto_version '1'", fmt.Sprintf("publication_names '%s'", w.cfg.PublicationName)}, // for pgoutput only
	})
	if err != nil {
		return err
	}

	// start reader node from last confirmed read offset
	//lastSlotReaderOffset := w.sysInfo.XLogPos
	w.walReaderOffset = w.slotConfirmedFlushLSN
	w.msgTypeMap = pgtype.NewMap()
	w.relationMsgMap = map[uint32]*pglogrepl.RelationMessage{}
	w.commitMsgTime = time.Now().Add(w.cfg.LogPollingInterval)
listenerLoop:
	for {
		select {
		case <-w.baseCtx.Done():
			break listenerLoop
		default:
		}

		errWorker := w.execLogListenerTask()
		if errWorker != nil && errors.Is(errWorker, agent.FatalError{}) {
			return errWorker
		} else if errWorker != nil {
			defaultLoggerWAL.Err(errWorker).Msg("worker process failed")
			continue
		}
	}
	return nil
}

func (w *WAL) commitLogOffset() error {
	if !time.Now().After(w.commitMsgTime) {
		return nil
	}

	errSendStatus := pglogrepl.SendStandbyStatusUpdate(w.baseCtx, w.conn, pglogrepl.StandbyStatusUpdate{
		WALWritePosition: w.walReaderOffset,
		WALFlushPosition: 0,
		WALApplyPosition: 0,
		ClientTime:       time.Time{},
		ReplyRequested:   false,
	})
	if errSendStatus != nil {
		return errSendStatus
	}
	defaultLoggerWAL.Info().Msg("standby status sent")
	w.commitMsgTime = time.Now().Add(w.cfg.LogPollingInterval)
	return nil
}

func (w *WAL) execLogListenerTask() error {
	w.inFlightProcesses.Add(1)
	defer func() {
		defaultLoggerWAL.Info().Msg("stopping log listener worker")
		w.inFlightProcesses.Done()
	}()
	defaultLoggerWAL.Info().
		Dur("polling_interval", w.cfg.LogPollingInterval).
		Dur("worker_timeout", w.cfg.WorkerTimeout).
		Msg("starting log listener worker")

	if errSendStatus := w.commitLogOffset(); errSendStatus != nil {
		return errSendStatus
	}

	// using background ctx to avoid in-flight process early stopping, thus not gracefully shutting down.
	scopedCtx, scopedCtxCancel := context.WithTimeout(context.Background(), w.cfg.WorkerTimeout)
	rawPsqlMsg, errRcvMsg := w.conn.ReceiveMessage(scopedCtx)
	defer scopedCtxCancel()
	if errRcvMsg != nil {
		if pgconn.Timeout(errRcvMsg) {
			return nil
		}
		return errRcvMsg
	}

	switch v := rawPsqlMsg.(type) {
	case *pgproto3.ErrorResponse:
		// ErrResponse is considered fatal due infinite looping capabilities during receive message process.
		// Thus, program termination is strictly required.
		return agent.FatalError{
			Err: errors.New(v.Message),
		}
	case *pgproto3.CopyData:
		defaultLoggerWAL.Info().Msg("received copy data")
		return w.processCopyData(v)
	default:
		defaultLoggerWAL.Warn().Msg("received unknown message")
	}
	return nil
}

func (w *WAL) processCopyData(cpData *pgproto3.CopyData) error {
	header := cpData.Data[0]
	data := cpData.Data[1:]

	switch header {
	case pglogrepl.PrimaryKeepaliveMessageByteID:
		pkm, errParse := pglogrepl.ParsePrimaryKeepaliveMessage(data)
		if errParse != nil {
			return errParse
		}

		defaultLoggerWAL.Info().
			Bool("reply_requested", pkm.ReplyRequested).
			Time("server_time", pkm.ServerTime).
			Uint64("server_wal_end", uint64(pkm.ServerWALEnd)).
			Msg("received primary keep-alive message")
		if pkm.ReplyRequested {
			w.commitMsgTime = time.Time{}
		}
	case pglogrepl.XLogDataByteID:
		xld, errParse := pglogrepl.ParseXLogData(data)
		if errParse != nil {
			return errParse
		}

		defaultLoggerWAL.Info().
			Time("server_time", xld.ServerTime).
			Uint64("wal_start", uint64(xld.WALStart)).
			Uint64("server_wal_end", uint64(xld.ServerWALEnd)).
			Int("wal_data_length", len(xld.WALData)).
			Msg("received x-log-data message")

		logicMsg, errLogParse := pglogrepl.Parse(xld.WALData)
		if errLogParse != nil {
			return errLogParse
		}

		if err := w.processLogicMessage(logicMsg); err != nil {
			return err
		}

		w.walReaderOffset = xld.WALStart + pglogrepl.LSN(len(xld.WALData))
		return nil
	}

	return nil
}

func decodeWALTextColumnData(mi *pgtype.Map, data []byte, dataType uint32) (interface{}, error) {
	if dt, ok := mi.TypeForOID(dataType); ok {
		return dt.Codec.DecodeValue(mi, dataType, pgtype.TextFormatCode, data)
	}
	return string(data), nil
}

func (w *WAL) processLogicMessage(msg pglogrepl.Message) error {
	switch logMsg := msg.(type) {
	case *pglogrepl.RelationMessage:
		w.relationMsgMap[logMsg.RelationID] = logMsg
	case *pglogrepl.InsertMessage:
		rel, ok := w.relationMsgMap[logMsg.RelationID]
		if !ok {
			return ErrWALUnknownRelation
		}

		values := map[string]any{}
		for idx, col := range logMsg.Tuple.Columns {
			colName := rel.Columns[idx].Name
			switch col.DataType {
			case 'n': // null
				values[colName] = nil
			case 'u': // unchanged toast
				// This TOAST value was not changed. TOAST values are not stored in the tuple, and logical replication doesn't want to spend a disk read to fetch its value for you.
			case 't': //text
				val, errDecode := decodeWALTextColumnData(w.msgTypeMap, col.Data, rel.Columns[idx].DataType)
				if errDecode != nil {
					return errDecode
				}
				values[colName] = val
			}
		}
		w.totalReads.Add(1)
		defaultLoggerWAL.Info().
			Str("namespace", rel.Namespace).
			Str("relation_name", rel.RelationName).
			Uint32("relation_id", rel.RelationID).
			Uint8("replica_identity", rel.ReplicaIdentity).
			Int("total_values", len(values)).
			Msg("decoded message")
		batch := egress.Batch{
			BatchID:           values["batch_id"].(string),
			TransportBatchRaw: values["raw_data"].([]byte),
			InsertTime:        values["insert_time"].(time.Time),
		}
		if fwdErr := w.fwd.ForwardBatch(batch); fwdErr != nil {
			defaultLoggerWAL.Err(fwdErr).Msg("forwarder failed to proxy message")
		}
	default:
		defaultLoggerWAL.Warn().
			Str("message_type", logMsg.Type().String()).
			Msg("received unknown logical message")
	}

	return nil
}

func (w *WAL) Close(ctx context.Context) error {
	defer func() {
		defaultLoggerWAL.Info().
			Uint64("total_reads", w.totalReads.Load()).
			Msg("listener successfully shut down")
	}()
	defaultLoggerWAL.Info().Msg("shutting down")
	w.baseCtxCancel()
	w.inFlightProcesses.Wait()
	if w.conn == nil {
		return nil
	}
	return w.conn.Close(ctx)
}
