package flows_test

import (
	"testing"
	"time"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/test"
	"github.com/nyaruka/goflow/utils/dates"
	"github.com/nyaruka/goflow/utils/uuids"

	"github.com/stretchr/testify/require"
)

func TestLegacyWebhookPayload(t *testing.T) {
	uuids.SetGenerator(uuids.NewSeededGenerator(123456))
	dates.SetNowSource(dates.NewSequentialNowSource(time.Date(2018, 7, 6, 12, 30, 0, 123456789, time.UTC)))
	defer uuids.SetGenerator(uuids.DefaultGenerator)
	defer dates.SetNowSource(dates.DefaultNowSource)

	session, _, err := test.CreateTestSession("", nil)
	run := session.Runs()[0]

	payload, err := run.EvaluateTemplate(flows.LegacyWebhookPayload)
	require.NoError(t, err)

	test.AssertEqualJSON(t, []byte(`{
		"channel": {
			"address": "+12345671111",
			"name": "My Android Phone",
			"uuid": "57f1078f-88aa-46f4-a59a-948a5739c03d"
		},
		"contact": {
			"name": "Ryan Lewis",
			"urn": "tel:+12065551212",
			"uuid": "5d76d86b-3bb9-4d5a-b822-c9d86f5d8e4f"
		},
		"flow": {
			"name": "Registration",
			"revision": 123,
			"uuid": "50c3706e-fedb-42c0-8eab-dda3335714b7"
		},
		"input": {
			"attachments": [
				{
					"content_type": "image/jpeg",
					"url": "http://s3.amazon.com/bucket/test.jpg"
				},
				{
					"content_type": "audio/mp3",
					"url": "http://s3.amazon.com/bucket/test.mp3"
				}
			],
			"channel": {
				"address": "+12345671111",
				"name": "My Android Phone",
				"uuid": "57f1078f-88aa-46f4-a59a-948a5739c03d"
			},
			"created_on": "2017-12-31T11:35:10.035757-02:00",
			"text": "Hi there",
			"type": "msg",
			"urn": {
				"display": "(206) 555-1212",
				"path": "+12065551212",
				"scheme": "tel"
			},
			"uuid": "9bf91c2b-ce58-4cef-aacc-281e03f69ab5"
		},
		"path": [
			{
				"arrived_on": "2018-07-06T12:30:03.123456Z",
				"exit_uuid": "d7a36118-0a38-4b35-a7e4-ae89042f0d3c",
				"node_uuid": "72a1f5df-49f9-45df-94c9-d86f7ea064e5",
				"uuid": "692926ea-09d6-4942-bd38-d266ec8d3716"
			},
			{
				"arrived_on": "2018-07-06T12:30:19.123456Z",
				"exit_uuid": "100f2d68-2481-4137-a0a3-177620ba3c5f",
				"node_uuid": "3dcccbb4-d29c-41dd-a01f-16d814c9ab82",
				"uuid": "5802813d-6c58-4292-8228-9728778b6c98"
			},
			{
				"arrived_on": "2018-07-06T12:30:28.123456Z",
				"exit_uuid": "d898f9a4-f0fc-4ac4-a639-c98c602bb511",
				"node_uuid": "f5bb9b7a-7b5e-45c3-8f0e-61b4e95edf03",
				"uuid": "970b8069-50f5-4f6f-8f41-6b2d9f33d623"
			},
			{
				"arrived_on": "2018-07-06T12:30:45.123456Z",
				"exit_uuid": "9fc5f8b4-2247-43db-b899-ab1ac50ba06c",
				"node_uuid": "c0781400-737f-4940-9a6c-1ec1c3df0325",
				"uuid": "5ecda5fc-951c-437b-a17e-f85e49829fb9"
			}
		],
		"results": {
			"2factor": {
				"category": "",
				"category_localized": "",
				"created_on": "2018-07-06T12:30:37.123456Z",
				"input": "",
				"name": "2Factor",
				"node_uuid": "f5bb9b7a-7b5e-45c3-8f0e-61b4e95edf03",
				"value": "34634624463525"
			},
			"favorite_color": {
				"category": "Red",
				"category_localized": "Red",
				"created_on": "2018-07-06T12:30:33.123456Z",
				"input": "",
				"name": "Favorite Color",
				"node_uuid": "f5bb9b7a-7b5e-45c3-8f0e-61b4e95edf03",
				"value": "red"
			},
			"phone_number": {
				"category": "",
				"category_localized": "",
				"created_on": "2018-07-06T12:30:29.123456Z",
				"input": "",
				"name": "Phone Number",
				"node_uuid": "f5bb9b7a-7b5e-45c3-8f0e-61b4e95edf03",
				"value": "+12344563452"
			}
		},
		"run": {
			"created_on": "2018-07-06T12:30:00.123456Z",
			"uuid": "d2f852ec-7b4e-457f-ae7f-f8b243c49ff5"
		}
	}`), []byte(payload), "payload mismatch")
}
