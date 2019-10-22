package migrations_test

import (
	"testing"

	"github.com/nyaruka/goflow/flows/definition/migrations"
	"github.com/nyaruka/goflow/utils/uuids"
)

func TestMigrate13_1(t *testing.T) {
	defer uuids.SetGenerator(uuids.DefaultGenerator)

	uuids.SetGenerator(uuids.NewSeededGenerator(123456))

	fn := migrations.AsParsed(migrations.Migrate13_1)

	original := `{
		"uuid": "76f0a02f-3b75-4b86-9064-e9195e1b3a02",
		"name": "Test Flow",
		"spec_version": "13.0",
		"language": "eng",
		"type": "messaging",
		"nodes": [
			{
				"uuid": "365293c7-633c-45bd-96b7-0b059766588d",
				"actions": [
					{
						"uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
						"type": "send_msg",
						"text": "Hi {1} of {2}, are you ready to complete today's survey?",
						"templating": {
							"template": {
								"uuid": "3ce100b7-a734-4b4e-891b-350b1279ade2",
								"name": "revive_issue"
							},
							"variables": [
								  "@contact.name", 
								  "@fields.state"
							]
						}
					}
				],
				"exits": [
					{
						"uuid": "b6f4caf3-ec99-44d5-a40c-8600ac0e2eac"
					}
				]
			}
		]
	}`
	expected := `{
		"uuid": "76f0a02f-3b75-4b86-9064-e9195e1b3a02",
		"name": "Test Flow",
		"spec_version": "13.0",
		"language": "eng",
		"type": "messaging",
		"nodes": [
			{
				"uuid": "365293c7-633c-45bd-96b7-0b059766588d",
				"actions": [
					{
						"uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
						"type": "send_msg",
						"text": "Hi {1} of {2}, are you ready to complete today's survey?",
						"templating": {
							"uuid": "d2f852ec-7b4e-457f-ae7f-f8b243c49ff5",
							"template": {
								"uuid": "3ce100b7-a734-4b4e-891b-350b1279ade2",
								"name": "revive_issue"
							},
							"variables": [
								  "@contact.name", 
								  "@fields.state"
							]
						}
					}
				],
				"exits": [
					{
						"uuid": "b6f4caf3-ec99-44d5-a40c-8600ac0e2eac"
					}
				]
			}
		]
	}`

	testMigration(t, fn, original, expected)
}
