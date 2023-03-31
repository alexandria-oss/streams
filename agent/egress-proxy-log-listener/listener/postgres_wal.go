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
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/viper"
)

var (
	errWALUnknownRelation = errors.New("unknown relation ID")
)

const WALDriver = "postgres_wal"

var defaultLoggerWAL = agent.DefaultLogger.With().
	Str("listener_driver", WALDriver).
	Logger()

type walConfig struct {
	OutputPlugin          string
	Username              string
	Password              string
	Address               string
	Port                  int
	Database              string
	UseSSL                bool
	SlotName              string
	EgressTable           string
	EnablePublishCreation bool
}

func newWALConfig() walConfig {
	viper.SetDefault("postgres.wal.slot_name", "streams_egress_proxy")
	viper.SetDefault("postgres.wal.egress_table", "streams_egress")
	viper.SetDefault("postgres.wal.enable_publish_creation", false)
	return walConfig{
		Username: viper.GetString("postgres.wal.username"),
		Password: viper.GetString("postgres.wal.password"),
		Address:  viper.GetString("postgres.wal.address"),
		Port:     viper.GetInt("postgres.wal.port"),
		Database: viper.GetString("postgres.wal.database"),
		UseSSL:   viper.GetBool("postgres.wal.use_ssl"),
		SlotName: viper.GetString("postgres.wal.slot_name"),
	}
}

func (c walConfig) connString() string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?replication=database", c.Username, c.Password, c.Address,
		c.Port, c.Database))
	if c.UseSSL {
		buf.WriteString("&sslMode=true")
	}

	return buf.String()
}

type WAL struct {
	cfg               walConfig
	baseCtx           context.Context
	baseCtxCancel     context.CancelFunc
	inFlightProcesses sync.WaitGroup
	totalReads        atomic.Uint64

	conn                  *pgconn.PgConn
	slotConfirmedFlushLSN pglogrepl.LSN
	sysInfo               pglogrepl.IdentifySystemResult
}

var _ Listener = &WAL{}

func NewWAL() *WAL {
	return &WAL{
		cfg:               newWALConfig(),
		inFlightProcesses: sync.WaitGroup{},
		totalReads:        atomic.Uint64{},
	}
}

func (w *WAL) Start() error {
	w.baseCtx, w.baseCtxCancel = context.WithCancel(context.Background())
	var err error
	w.conn, err = pgconn.Connect(w.baseCtx, w.cfg.connString())
	if err != nil {
		return err
	}

	if err = w.createPublication(); err != nil {
		return err
	}

	if err = w.fetchSysInfo(); err != nil {
		return err
	}

	if err = w.fetchSlotInfo(); err != nil {
		return err
	}

	return w.listenLogs()
}

func (w *WAL) createPublication() error {
	if !w.cfg.EnablePublishCreation {
		return nil
	}
	query := fmt.Sprintf("CREATE PUBLICATION %s FOR TABLE %s;", w.cfg.SlotName, w.cfg.EgressTable)
	result := w.conn.Exec(w.baseCtx, query)
	_, err := result.ReadAll()
	if err != nil {
		return err
	}
	defaultLoggerWAL.Info().
		Str("publication_name", w.cfg.SlotName).
		Str("table_name", w.cfg.EgressTable).
		Msg("created publication for specified table")
	return nil
}

func (w *WAL) createSlot() error {
	if !w.cfg.EnablePublishCreation {
		return nil
	}
	query := fmt.Sprintf("CREATE PUBLICATION %s FOR TABLE %s;", w.cfg.SlotName, w.cfg.EgressTable)
	result := w.conn.Exec(w.baseCtx, query)
	_, err := result.ReadAll()
	if err != nil {
		return err
	}
	defaultLoggerWAL.Info().
		Str("publication_name", w.cfg.SlotName).
		Str("table_name", w.cfg.EgressTable).
		Msg("created publication for specified table")
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
	defer res.Close()

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

func decodeWALTextColumnData(mi *pgtype.Map, data []byte, dataType uint32) (interface{}, error) {
	if dt, ok := mi.TypeForOID(dataType); ok {
		return dt.Codec.DecodeValue(mi, dataType, pgtype.TextFormatCode, data)
	}
	return string(data), nil
}

func (w *WAL) listenLogs() error {
	err := pglogrepl.StartReplication(w.baseCtx, w.conn,
		w.cfg.SlotName,
		w.slotConfirmedFlushLSN,
		pglogrepl.StartReplicationOptions{
			PluginArgs: []string{"proto_version '1'", fmt.Sprintf("publication_names '%s'", w.cfg.SlotName)}, // for pgoutput only
		})
	if err != nil {
		return err
	}

	// TODO: Add last read offset for a replication slot (slot = instance app node), so node can start processing were it left off before shutdown or even at fresh start
	// TODO: Add worker_id or similar to egress_table so
	clientXLogPos := w.slotConfirmedFlushLSN
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)
	relations := map[uint32]*pglogrepl.RelationMessage{}
	typeMap := pgtype.NewMap()

mainLoop:
	for {
		if time.Now().After(nextStandbyMessageDeadline) {
			err = pglogrepl.SendStandbyStatusUpdate(w.baseCtx, w.conn,
				pglogrepl.StandbyStatusUpdate{
					WALWritePosition: clientXLogPos,
				})
			if err != nil {
				defaultLoggerWAL.Err(err).Msg("listener processing error")
				continue mainLoop
			}
			defaultLoggerWAL.Info().Msg("sent standby status message")
			nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		}

		scopedCtx, cancel := context.WithDeadline(w.baseCtx, nextStandbyMessageDeadline)
		rawMsg, errReceive := w.conn.ReceiveMessage(scopedCtx)
		cancel()
		if errReceive != nil {
			if pgconn.Timeout(errReceive) {
				continue mainLoop
			}
			defaultLoggerWAL.Err(err).Msg("listener processing error")
			continue mainLoop
		}

		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			defaultLoggerWAL.Err(errors.New(errMsg.Message)).Msg("listener processing error")
			continue mainLoop
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			defaultLoggerWAL.Warn().Msg("received unexpected message")
			continue mainLoop
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, errParse := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if errParse != nil {
				defaultLoggerWAL.Err(errParse).Msg("listener processing error")
				continue mainLoop
			}
			defaultLoggerWAL.Info().
				Str("server_wal_end", pkm.ServerWALEnd.String()).
				Bool("reply_requested", pkm.ReplyRequested).
				Time("server_time", pkm.ServerTime).
				Msg("received primary keep alive message")

			if pkm.ReplyRequested {
				nextStandbyMessageDeadline = time.Time{}
			}
		case pglogrepl.XLogDataByteID:
			logData, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				defaultLoggerWAL.Err(err).Msg("listener processing error")
				continue mainLoop
			}

			logicalMsgInterface, err := pglogrepl.Parse(logData.WALData)
			if err != nil {
				defaultLoggerWAL.Err(err).Msg("listener processing error")
				continue mainLoop
			}

			switch logicalMsg := logicalMsgInterface.(type) {
			case *pglogrepl.RelationMessage:
				relations[logicalMsg.RelationID] = logicalMsg
			case *pglogrepl.InsertMessage:
				w.totalReads.Add(1)
				defaultLoggerWAL.Info().
					Uint64("total_reads", w.totalReads.Load()).
					Interface("data", logicalMsg.Tuple.Columns).
					Msg("received data")
				rel, ok := relations[logicalMsg.RelationID]
				if !ok {
					defaultLoggerWAL.Err(errWALUnknownRelation).Msg("listener processing error")
					continue mainLoop
				}

				values := map[string]any{}
				for idx, col := range logicalMsg.Tuple.Columns {
					colName := rel.Columns[idx].Name
					switch col.DataType {
					case 'n': // null
						values[colName] = nil
					case 'u': // unchanged toast
						// This TOAST value was not changed. TOAST values are not stored in the tuple, and logical replication doesn't want to spend a disk read to fetch its value for you.
					case 't': //text
						val, err := decodeWALTextColumnData(typeMap, col.Data, rel.Columns[idx].DataType)
						if err != nil {
							defaultLoggerWAL.Err(err).Msg("listener processing error")
							continue mainLoop
						}
						values[colName] = val
					}
				}
				defaultLoggerWAL.Info().
					Str("namespace", rel.Namespace).
					Str("relation_name", rel.RelationName).
					Interface("values", values).
					Msg("decoded message")
			default:
				defaultLoggerWAL.Warn().Msg("unknown message type in pgoutput stream")
			}

			clientXLogPos = logData.WALStart + pglogrepl.LSN(len(logData.WALData))
		}
		select {
		case <-w.baseCtx.Done():
			break mainLoop
		default:
		}
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
	//w.inFlightProcesses.Wait()
	w.baseCtxCancel()
	return w.conn.Close(ctx)
}
