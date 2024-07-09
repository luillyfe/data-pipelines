package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"google.golang.org/api/option"
)

// FirestoreWriter is a PTransform that writes data to Firestore
type FirestoreWriter struct {
	ProjectID  string
	Collection string
	CredPath   string

	client     *firestore.Client
	bulkWriter *firestore.BulkWriter
}

func (f *FirestoreWriter) ProcessElement(ctx context.Context, question *MultipleChoiceQuestion) error {
	// Lazy initialization
	if f.client == nil {
		if err := f.setupClient(ctx); err != nil {
			return err
		}
	}

	// The Firestore's NewDoc() method generates a new document with an auto-generated ID
	ref := f.client.Collection(f.Collection).NewDoc()
	_, err := f.bulkWriter.Create(ref, question)
	if err != nil {
		log.Printf("Error queuing write operation: %v", err)
		return err
	}

	return nil
}

// This registration is crucial for serialization purposes in distributed processing environments.
func init() {
	// 2 inputs and 1 output => DoFn2x1
	// Type arguments [context.Context, *MultipleChoiceQuestion, error]
	register.DoFn2x1(&FirestoreWriter{})
}

func (f *FirestoreWriter) setupClient(ctx context.Context) error {
	client, err := firestore.NewClient(ctx, f.ProjectID, option.WithCredentialsFile(f.CredPath))
	if err != nil {
		return err
	}

	f.client = client
	f.bulkWriter = client.BulkWriter(ctx)
	return nil
}

func (f *FirestoreWriter) Teardown() error {
	if f.client != nil {
		return f.client.Close()
	}
	return nil
}

func (f *FirestoreWriter) FinishBundle(ctx context.Context) {
	if f.bulkWriter != nil {
		f.bulkWriter.Flush()
	}
}
