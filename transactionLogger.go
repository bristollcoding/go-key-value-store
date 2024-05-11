package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
)

//Transaction logger allows store to get back to its previous state redoing every action until its last state

// Interface for generating transaction log (will implement this differently depending on persistence)
type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)

	Err() <-chan error
	ReadTransLog() (<-chan Event, <-chan error)
	Start()
}

// Type to transport transaction information ( Event)
type Event struct {
	Sequence  uint64 //transaction id
	EventType EventType
	Key       string
	Value     string
}
type EventType string

const (
	//Codes for loggable transactions
	EventDelete EventType = "D"
	EventPut    EventType = "P"
)

// Transaction logger with file as persistence
type FileTransactionLogger struct {
	//Channel for events ( buffered)
	events chan<- Event
	//Channel for error
	errors <-chan error
	//Last used transaction id (sequence)
	lastSequence uint64
	//File location
	file *os.File

	wg *sync.WaitGroup
}

var tLogger *FileTransactionLogger

// Create a FileTransactionLogger
func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	var err error
	//Create Transaction Logger
	tLog := FileTransactionLogger{wg: &sync.WaitGroup{}}

	//Opens file in read/write mode and assign it to logger, write always append and if not exists create file
	tLog.file, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)

	if err != nil {
		fmt.Println("6")
		return nil, fmt.Errorf("error opening transaction log file: %v with error: %w ", filename, err)
	}

	return &tLog, nil
}

// Function to recreate transactions from log file
func initTransLogger() error {

	var err error

	//create new logger
	tLogger, err = NewFileTransactionLogger("transLog.log")

	if err != nil {
		fmt.Println("5")
		return fmt.Errorf("error crating file transaction logger: %w", err)
	}
	eventsChan, errorsChan := tLogger.ReadTransLog()

	e := Event{}
	ok := true

	for ok && err == nil {
		fmt.Println("reading from channels inside init")
		// Read from channesl
		select {

		case err, ok = <-errorsChan:
		case e, ok = <-eventsChan:
			switch e.EventType {
			case EventDelete:
				fmt.Println("delete event")
				Delete(e.Key)
			case EventPut:
				fmt.Println("put event")

				err = Put(e.Key, e.Value)
			}
		}
		fmt.Println("End reading from channels init")
	}
	tLogger.wg.Wait()
	if tLogger.events != nil {
		close(tLogger.events) // Terminates Run loop and goroutine
	}
	tLogger.Start()

	//Get errors from error channel
	go func() {
		for err := range tLogger.Err() {
			log.Print(err)
		}
	}()
	return err
}

// Function to start waiting for log events
func (l *FileTransactionLogger) Start() {
	fmt.Println("Transsaction Logger Start")
	// Initialize channels
	events := make(chan Event, 10) //Max 10 events queued
	l.events = events
	errors := make(chan error, 1)
	l.errors = errors

	// Goroutine
	go func() {
		for e := range events {
			//Increase last transactionID
			l.lastSequence++
			//Transaction format in file-> [id TransactionType Key Value] each in new line
			//TODO: handle values with formats (\n,\t ...)
			wb, err := fmt.Fprintf(l.file, "%d\t%s\t%s\t%s\n", l.lastSequence, e.EventType, e.Key, e.Value)
			if err != nil {
				fmt.Println("4")
				//Insert err into erros channel and exit
				errors <- fmt.Errorf("cannot write to log file: %w", err)

			}
			fmt.Printf("%d bytes written to file: %s \n", wb, l.file.Name())

			l.wg.Done()
		}
	}()
}

// Function to read from log file --> Read concurrently using channels because only final state matters
func (l *FileTransactionLogger) ReadTransLog() (<-chan Event, <-chan error) {

	//Buff scanner to read file
	scanner := bufio.NewScanner(l.file)
	//Create out channels
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	//Goroutine to read from file and write to outEvent channel
	go func() {
		var e Event

		// make sure channels are closed on exit
		defer close(outError)
		defer close(outEvent)

		//Read log file
		for scanner.Scan() {

			line := scanner.Text()

			//Read and parse event from line
			wb, err := fmt.Sscanf(line, "%d\t%s\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				fmt.Println("3")
				outError <- fmt.Errorf("parse error reading from log file: %w", err)
				return
			}
			//Only for debug purposes
			fmt.Printf("%d items parsed correctly\n", wb)
			fmt.Printf("LastSeq: %d eSeq: %d\n", l.lastSequence, e.Sequence)
			//Check if sequence is in correct order (increasing), else something is wrong with file
			if l.lastSequence >= e.Sequence {
				fmt.Println("2")
				outError <- fmt.Errorf("transaction sequence out of order")
				return
			}
			// uv, err := url.QueryUnescape(e.Value)
			// if err != nil {
			// 	fmt.Println("1")
			// 	outError <- fmt.Errorf("value decoding failure: %w", err)
			// 	return
			// }

			// e.Value = uv
			//Set lastSequence to last readed event sequence number
			l.lastSequence = e.Sequence

			//Send event to channel
			outEvent <- e
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("7")
			outError <- fmt.Errorf("transaction log scanner failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}

func (l *FileTransactionLogger) Wait() {
	l.wg.Wait()
}

func (l *FileTransactionLogger) Close() error {
	l.wg.Wait()

	if l.events != nil {
		close(l.events) // Terminates Run loop and goroutine
	}

	return l.file.Close()
}

// Implements TransactionLogger interface for FileTransactionLogger

func (l *FileTransactionLogger) WriteDelete(key string) {
	//add to the waiting group
	l.wg.Add(1)
	// write Delete Event to events channel
	l.events <- Event{EventType: EventType(EventDelete), Key: key}
}
func (l *FileTransactionLogger) WritePut(key, value string) {
	fmt.Println("Start WritePut with key: " + key)
	//add to the waiting group
	l.wg.Add(1)
	//Write Put event to events channel
	l.events <- Event{EventType: EventType(EventPut), Key: key, Value: value}
	fmt.Println("End WritePut")
}

// func access erros channel
func (l *FileTransactionLogger) Err() <-chan error {

	return l.errors
}
