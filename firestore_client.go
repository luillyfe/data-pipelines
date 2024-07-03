package main

import (
	"context"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"google.golang.org/api/option"
)

// https://firebase.google.com/docs/firestore/manage-data/transactions
const (
	maxBatchSize = 500 // A batched write can contain up to 500 operations.
	maxBatchTime = 1 * time.Second
)

// FirestoreWriter is a PTransform that writes data to Firestore
type FirestoreWriter struct {
	ProjectID  string
	Collection string

	count      int
	mu         sync.Mutex
	client     *firestore.Client
	bulkWriter *firestore.BulkWriter
	lastWrite  time.Time
}

func (f *FirestoreWriter) ProcessElement(ctx context.Context, question *Question) error {
	// Lock the mutex to ensure thread-safety
	// This is necessary because ProcessElement might be called concurrently
	f.mu.Lock()
	defer f.mu.Unlock()

	// The Firestore's NewDoc() method generates a new document with an auto-generated ID
	ref := f.client.Collection(f.Collection).NewDoc()
	_, err := f.bulkWriter.Create(ref, question)
	if err != nil {
		log.Printf("Error queuing write operation: %v", err)
		return err
	}

	// This keeps track of how many operations we've queued.
	f.count++

	// Flexible flushing mechanism. Based on count: When we reach maxBatchSize operations.
	// Based on time: When maxBatchTime has elapsed since the last flush.
	// This ensures that we're balancing between efficient batching and timely writes.
	if f.count >= maxBatchSize || time.Since(f.lastWrite) >= maxBatchTime {
		f.bulkWriter.Flush()
	}

	return nil
}

// Called once per bundle
func (f *FirestoreWriter) Setup(ctx context.Context) error {
	// initialize the Firestore client
	client, err := firestore.NewClient(ctx, f.ProjectID, option.WithCredentialsFile(".credentials/firestore.json"))
	if err != nil {
		return err
	}

	// Firestore Client, BulkWriter instance
	f.client = client
	f.bulkWriter = client.BulkWriter(ctx)
	f.lastWrite = time.Now()
	return nil
}

func (f *FirestoreWriter) Teardown() error {
	if f.client != nil {
		return f.client.Close()
	}
	return nil
}

func (f *FirestoreWriter) FinishBundle() {
	f.bulkWriter.Flush()
}

// This registration is crucial for serialization purposes in distributed processing environments.
func init() {
	// 2 inputs and 1 output => DoFn2x1
	// Type arguments [context.Context, *Question, error]
	register.DoFn2x1(&FirestoreWriter{})
}
