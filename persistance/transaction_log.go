package persistance

import (
	"bufio"
	"fmt"
	"os"
)

type EventType byte

const (
	_                     = iota // ignore 0
	EventPut    EventType = iota
	EventDelete EventType = iota
)

type TransactionLogger interface {
	// for putting a PUT event inside our log
	WritePut(key, value string)
	// for putting a DELETE event inside our log
	WriteDelete(key string)
	// method for reading events written by logger
	ReadEvents() (<-chan Event, <-chan error)
	// function that returns read only channel for reading errors
	Err() <-chan error
	// method for running the logger
	Run()
}

// represents an event that happened in the store
type Event struct {
	Sequence uint64
	Type     EventType
	Key      string
	Value    string
}

type FileTransactionalLogger struct {
	events       chan<- Event // write only channel for appending events
	errors       <-chan error // read only channel for reading errors
	lastSequence uint64       // last used Sequence num in log
	file         *os.File     // physical location of log
}

// function accepts as parameter type of logger to create and the creates it
// in case of file transaction logger filename must be set as an env param (TLOG_FILENAME)
func NewTransactionLogger(logger string) (TransactionLogger, error) {
	switch logger {
	case "file":
		return newTransactionFileLogger(os.Getenv("TLOG_FILENAME"))
	case "":
		return nil, fmt.Errorf("transaction logger type not defined")
	default:
		return nil, fmt.Errorf("no such transaction logger %s", logger)
	}
}

// function for creating a file transaction logger
func newTransactionFileLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &FileTransactionalLogger{file: file}, nil
}

func (l *FileTransactionalLogger) Run() {
	events := make(chan Event, 12)
	l.events = events
	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {
			l.lastSequence++

			_, err := fmt.Fprintf(l.file, "%d\t%d\t%s\t%s\n", l.lastSequence, e.Type, e.Key, e.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (l *FileTransactionalLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file)

	outEvents := make(chan Event)
	outError := make(chan error)

	go func() {
		var event Event
		defer close(outEvents)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &event.Sequence, &event.Type, &event.Key, &event.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}

			// check if sequence nums are in order
			if l.lastSequence >= event.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			l.lastSequence = event.Sequence // update last sequence
			outEvents <- event
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvents, outError
}

func (l *FileTransactionalLogger) Err() <-chan error {
	return l.errors
}

func (l *FileTransactionalLogger) WritePut(key, value string) {
	l.events <- Event{Type: EventPut, Key: key, Value: value}
}

func (l *FileTransactionalLogger) WriteDelete(key string) {
	l.events <- Event{Type: EventDelete, Key: key}
}
