// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigquery

import (
	"testing"
	"time"

	"cloud.google.com/go/internal/testutil"

	bq "google.golang.org/api/bigquery/v2"
)

func TestBQTableToMetadata(t *testing.T) {
	aTime := time.Date(2017, 1, 26, 0, 0, 0, 0, time.Local)
	aTimeMillis := aTime.UnixNano() / 1e6
	for _, test := range []struct {
		in   *bq.Table
		want *TableMetadata
	}{
		{&bq.Table{}, &TableMetadata{}}, // test minimal case
		{
			&bq.Table{
				CreationTime:     aTimeMillis,
				Description:      "desc",
				Etag:             "etag",
				ExpirationTime:   aTimeMillis,
				FriendlyName:     "fname",
				Id:               "id",
				LastModifiedTime: uint64(aTimeMillis),
				Location:         "loc",
				NumBytes:         123,
				NumLongTermBytes: 23,
				NumRows:          7,
				StreamingBuffer: &bq.Streamingbuffer{
					EstimatedBytes:  11,
					EstimatedRows:   3,
					OldestEntryTime: uint64(aTimeMillis),
				},
				TimePartitioning: &bq.TimePartitioning{
					ExpirationMs: 7890,
					Type:         "DAY",
				},
				Type: "EXTERNAL",
				View: &bq.ViewDefinition{Query: "view-query"},
			},
			&TableMetadata{
				Description:      "desc",
				Name:             "fname",
				View:             "view-query",
				ID:               "id",
				Type:             ExternalTable,
				ExpirationTime:   aTime.Truncate(time.Millisecond),
				CreationTime:     aTime.Truncate(time.Millisecond),
				LastModifiedTime: aTime.Truncate(time.Millisecond),
				NumBytes:         123,
				NumRows:          7,
				TimePartitioning: &TimePartitioning{Expiration: time.Duration(7890) * time.Millisecond},
				StreamingBuffer: &StreamingBuffer{
					EstimatedBytes:  11,
					EstimatedRows:   3,
					OldestEntryTime: aTime,
				},
				ETag: "etag",
			},
		},
	} {
		got := bqTableToMetadata(test.in)
		if !testutil.Equal(got, test.want) {
			t.Errorf("%v:\ngot  %+v\nwant %+v", test.in, got, test.want)
		}
	}
}

func TestBQDatasetFromMetadata(t *testing.T) {
	for _, test := range []struct {
		in   *DatasetMetadata
		want *bq.Dataset
	}{
		{nil, &bq.Dataset{}},
		{&DatasetMetadata{Name: "name"}, &bq.Dataset{FriendlyName: "name"}},
		{&DatasetMetadata{
			Name:                   "name",
			Description:            "desc",
			DefaultTableExpiration: time.Hour,
			Location:               "EU",
			Labels:                 map[string]string{"x": "y"},
		}, &bq.Dataset{
			FriendlyName:             "name",
			Description:              "desc",
			DefaultTableExpirationMs: 60 * 60 * 1000,
			Location:                 "EU",
			Labels:                   map[string]string{"x": "y"},
		}},
	} {
		got, err := bqDatasetFromMetadata(test.in)
		if err != nil {
			t.Fatal(err)
		}
		if !testutil.Equal(got, test.want) {
			t.Errorf("%v:\ngot  %+v\nwant %+v", test.in, got, test.want)
		}
	}

	// Check that non-writeable fields are unset.
	_, err := bqDatasetFromMetadata(&DatasetMetadata{FullID: "x"})
	if err == nil {
		t.Error("got nil, want error")
	}
}

func TestBQDatasetFromUpdateMetadata(t *testing.T) {
	dm := DatasetMetadataToUpdate{
		Description: "desc",
		Name:        "name",
		DefaultTableExpiration: time.Hour,
	}
	dm.SetLabel("label", "value")
	dm.DeleteLabel("del")

	got := bqDatasetFromUpdateMetadata(&dm)
	want := &bq.Dataset{
		Description:              "desc",
		FriendlyName:             "name",
		DefaultTableExpirationMs: 60 * 60 * 1000,
		Labels:          map[string]string{"label": "value"},
		ForceSendFields: []string{"Description", "FriendlyName"},
		NullFields:      []string{"Labels.del"},
	}
	if diff := testutil.Diff(got, want); diff != "" {
		t.Errorf("-got, +want:\n%s", diff)
	}
}
