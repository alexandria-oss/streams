package main

import (
	"context"
	"encoding/hex"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	agent "github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener"

	"github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener/listener"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("streams")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	driver := viper.GetString("agent.driver")
	l, err := listener.NewListener(driver)
	if err != nil {
		panic(err)
	}

	agent.DefaultLogger.Info().
		Str("listener_driver", driver).
		Msg("driver detected")

	sysChan := make(chan os.Signal, 2)
	signal.Notify(sysChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		if err = l.Start(); err != nil {
			agent.DefaultLogger.Err(err).Msg("listener execution process failed")
			os.Exit(1)
		}
	}()
	<-sysChan

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err = l.Close(shutdownCtx); err != nil {
		agent.DefaultLogger.Err(err).Msg("listener graceful shutdown process failed")
		return
	}
	agent.DefaultLogger.Info().Msg("agent shutdown succeeded")
}

func readWAL() {
	const outputPlugin = "pgoutput"
	conn, err := pgconn.Connect(context.Background(), "postgres://streams_egress_proxy_replicator:foobar@127.0.0.1/sample_database?replication=database")
	if err != nil {
		log.Fatalln("failed to connect to PostgreSQL server:", err)
	}
	defer conn.Close(context.Background())

	var pluginArguments []string
	if outputPlugin == "pgoutput" {
		pluginArguments = []string{"proto_version '1'", "publication_names 'pglogrepl_demo'"}
	} else if outputPlugin == "wal2json" {
		pluginArguments = []string{"\"pretty-print\" 'true'"}
	}

	sysident, err := pglogrepl.IdentifySystem(context.Background(), conn)
	if err != nil {
		log.Fatalln("IdentifySystem failed:", err)
	}
	log.Println("SystemID:", sysident.SystemID, "Timeline:", sysident.Timeline, "XLogPos:", sysident.XLogPos, "DBName:", sysident.DBName)

	slotName := "streams_egress_proxy"

	err = pglogrepl.StartReplication(context.Background(), conn, slotName, sysident.XLogPos, pglogrepl.StartReplicationOptions{PluginArgs: pluginArguments})
	if err != nil {
		log.Fatalln("StartReplication failed:", err)
	}
	log.Println("Logical replication started on slot", slotName)

	clientXLogPos := sysident.XLogPos
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)
	relations := map[uint32]*pglogrepl.RelationMessage{}
	typeMap := pgtype.NewMap()

	for {
		if time.Now().After(nextStandbyMessageDeadline) {
			err = pglogrepl.SendStandbyStatusUpdate(context.Background(), conn, pglogrepl.StandbyStatusUpdate{WALWritePosition: clientXLogPos})
			if err != nil {
				log.Fatalln("SendStandbyStatusUpdate failed:", err)
			}
			log.Println("Sent Standby status message")
			nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		}

		ctx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
		rawMsg, err := conn.ReceiveMessage(ctx)
		cancel()
		if err != nil {
			if pgconn.Timeout(err) {
				continue
			}
			log.Fatalln("ReceiveMessage failed:", err)
		}

		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			log.Fatalf("received Postgres WAL error: %+v", errMsg)
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			log.Printf("Received unexpected message: %T\n", rawMsg)
			continue
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				log.Fatalln("ParsePrimaryKeepaliveMessage failed:", err)
			}
			log.Println("Primary Keepalive Message =>", "ServerWALEnd:", pkm.ServerWALEnd, "ServerTime:", pkm.ServerTime, "ReplyRequested:", pkm.ReplyRequested)

			if pkm.ReplyRequested {
				nextStandbyMessageDeadline = time.Time{}
			}

		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				log.Fatalln("ParseXLogData failed:", err)
			}
			log.Printf("XLogData => WALStart %s ServerWALEnd %s ServerTime %s WALData:\n%s\n", xld.WALStart, xld.ServerWALEnd, xld.ServerTime, hex.Dump(xld.WALData))
			logicalMsg, err := pglogrepl.Parse(xld.WALData)
			if err != nil {
				log.Fatalf("Parse logical replication message: %s", err)
			}
			log.Printf("Receive a logical replication message: %s", logicalMsg.Type())
			switch logicalMsg := logicalMsg.(type) {
			case *pglogrepl.RelationMessage:
				relations[logicalMsg.RelationID] = logicalMsg

			case *pglogrepl.BeginMessage:
				// Indicates the beginning of a group of changes in a transaction. This is only sent for committed transactions. You won't get any events from rolled back transactions.

			case *pglogrepl.CommitMessage:

			case *pglogrepl.InsertMessage:
				rel, ok := relations[logicalMsg.RelationID]
				if !ok {
					log.Fatalf("unknown relation ID %d", logicalMsg.RelationID)
				}
				values := map[string]interface{}{}
				for idx, col := range logicalMsg.Tuple.Columns {
					colName := rel.Columns[idx].Name
					switch col.DataType {
					case 'n': // null
						values[colName] = nil
					case 'u': // unchanged toast
						// This TOAST value was not changed. TOAST values are not stored in the tuple, and logical replication doesn't want to spend a disk read to fetch its value for you.
					case 't': //text
						val, err := decodeTextColumnData(typeMap, col.Data, rel.Columns[idx].DataType)
						if err != nil {
							log.Fatalln("error decoding column data:", err)
						}
						values[colName] = val
					}
				}
				log.Printf("INSERT INTO %s.%s: %v", rel.Namespace, rel.RelationName, values)

			case *pglogrepl.UpdateMessage:
				// ...
			case *pglogrepl.DeleteMessage:
				// ...
			case *pglogrepl.TruncateMessage:
				// ...

			case *pglogrepl.TypeMessage:
			case *pglogrepl.OriginMessage:
			default:
				log.Printf("Unknown message type in pgoutput stream: %T", logicalMsg)
			}

			clientXLogPos = xld.WALStart + pglogrepl.LSN(len(xld.WALData))
		}
	}
}

func decodeTextColumnData(mi *pgtype.Map, data []byte, dataType uint32) (interface{}, error) {
	if dt, ok := mi.TypeForOID(dataType); ok {
		return dt.Codec.DecodeValue(mi, dataType, pgtype.TextFormatCode, data)
	}
	return string(data), nil
}
