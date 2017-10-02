package consumer

import (
	"context"
	"reflect"
	"testing"

	logt "github.com/sirupsen/logrus/hooks/test"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

func Test_handleMetadataCreateRequest_errMessageType(t *testing.T) {
	c, _ := getConsumer(t)
	if err := c.handleMetadataCreateRequest(&message.Message{}); err == nil {
		t.Fatal("Expected non-nil error, got nil")
	}
}
func Test_handleMetadataCreateRequest_emptyFiles(t *testing.T) {
	c, _ := getConsumer(t)
	msg := message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	if err := c.handleMetadataCreateRequest(msg); err != nil {
		t.Fatalf("Expected nil error, got %s", err)
	}
}

func Test_getFilename(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"https://foo12345.com/foobar.jpg", "foobar.jpg"},
		{"https://foo12345.com/foo/bar/rrr", "foo/bar/rrr"},
		{":invalid-url:", ""},
	}
	for _, tt := range tests {
		if got := getFilename(tt.path); tt.want != got {
			t.Errorf("getFilename(); want %v, got %v", tt.want, got)
		}
	}
}

func Test_describeDataset(t *testing.T) {
	var (
		ts   = getTransferSession(t)
		want = map[string][][2]string{
			"objects/": [][2]string{
				[2]string{"dc.title", "Title"},
				[2]string{"dc.type", "audio"},
				[2]string{"dc.identifier", "I1"},
				[2]string{"dc.identifier", "I2"},
				[2]string{"dcterms.issued", "date of publication"},
				[2]string{"dc.publisher", "orgname"},
				[2]string{"dc.contributor", "person 2"},
			},
		}
	)
	describeDataset(ts, &message.MetadataCreateRequest{
		message.ResearchObject{
			ObjectTitle:        "Title",
			ObjectResourceType: message.ResourceTypeEnum_audio,
			ObjectIdentifier: []message.Identifier{
				message.Identifier{IdentifierValue: "I1"},
				message.Identifier{IdentifierValue: "I2"},
			},
			ObjectDate: []message.Date{
				message.Date{DateType: message.DateTypeEnum_published, DateValue: "date of publication"},
				message.Date{DateType: message.DateTypeEnum_accepted, DateValue: "date of ..."},
			},
			ObjectOrganisationRole: []message.OrganisationRole{
				message.OrganisationRole{
					Organisation: &message.Organisation{OrganisationName: "orgname"},
				},
			},
			ObjectPersonRole: []message.PersonRole{
				message.PersonRole{Role: message.PersonRoleEnum_relatedPerson, Person: &message.Person{PersonGivenName: "person 1"}},
				message.PersonRole{Role: message.PersonRoleEnum_dataCreator, Person: &message.Person{PersonGivenName: "person 2"}},
			},
		},
	})
	entries := ts.Metadata.Entries()
	if !reflect.DeepEqual(want, entries) {
		t.Fatalf("describeFile(); unexpected result, want %v, got %v", want, entries)
	}
}

func Test_describeFile(t *testing.T) {
	var (
		ts   = getTransferSession(t)
		want = map[string][][2]string{
			"objects/foobar.jpg": [][2]string{
				[2]string{"dc.identifier", "I1"},
				[2]string{"dc.title", "foobar.jpg"},
			},
		}
	)
	describeFile(ts, "foobar.jpg", &message.File{FileIdentifier: "I1", FileName: "foobar.jpg"})
	entries := ts.Metadata.Entries()
	if !reflect.DeepEqual(want, entries) {
		t.Fatalf("describeFile(); unexpected result, want %v, got %v", want, entries)
	}
}

func getConsumer(t *testing.T) (*ConsumerImpl, *logt.Hook) {
	logger, hook := logt.NewNullLogger()
	c := &ConsumerImpl{
		ctx:    context.Background(),
		logger: logger,
	}
	return c, hook
}

func getTransferSession(t *testing.T) *amclient.TransferSession {
	return &amclient.TransferSession{
		Metadata: amclient.NewMetadataSet(nil),
	}
}
