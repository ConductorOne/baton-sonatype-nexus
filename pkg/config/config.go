package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	// Add the SchemaFields for the Config.
	configField = field.StringField("configField")
	HostField   = field.StringField("host",
		field.WithDescription("Nexus host URL"),
		field.WithDefaultValue("http://localhost:8081"),
		field.WithRequired(true))
	UsernameField = field.StringField("username",
		field.WithDescription("Nexus username"),
		field.WithRequired(true))
	PasswordField = field.StringField("password",
		field.WithDescription("Nexus password"),
		field.WithRequired(true))
	ConfigurationFields = []field.SchemaField{configField, HostField, UsernameField, PasswordField}

	// FieldRelationships defines relationships between the ConfigurationFields that can be automatically validated.
	// For example, a username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run -tags=generate ./gen
var Config = field.NewConfiguration(
	ConfigurationFields,
	field.WithConstraints(FieldRelationships...),
	field.WithConnectorDisplayName("Sonatype Nexus"),
)
