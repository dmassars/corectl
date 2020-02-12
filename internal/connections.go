package internal

import (
	"context"
	"github.com/qlik-oss/corectl/internal/log"
	"github.com/qlik-oss/corectl/pkg/boot"
	"github.com/qlik-oss/enigma-go"
)

// SetupConnections reads all connections from both the project file path and the config file path and updates
// the list of connections in the app.
func SetupConnections(ctx context.Context, doc *enigma.Doc, config *boot.ConnectionsConfig) error {

	connections, err := doc.GetConnections(ctx)

	if config == nil || config.Connections == nil {
		return nil
	}

	connectionConfigEntries := *config.Connections

	for name, configEntry := range connectionConfigEntries {
		var connection = &enigma.Connection{
			Name:     name,
			Type:     configEntry.Type,
			UserName: configEntry.Username,
			Password: configEntry.Password,
		}

		if configEntry.ConnectionString != "" {
			connection.ConnectionString = configEntry.ConnectionString
		} else {
			connection.ConnectionString = "CUSTOM CONNECT TO \"provider=" + configEntry.Type + ";" + flattenSettings(configEntry.Settings) + "\""
		}

		if existingConnectionID := findExistingConnection(connections, connection.Name); existingConnectionID != "" {
			log.Verboseln("Modifying connection: " + connection.Name + " (" + existingConnectionID + ")")
			err = doc.ModifyConnection(ctx, existingConnectionID, connection, true)
		} else {
			log.Verboseln("Creating new connection: " + connection.Name)
			var id string
			id, err = doc.CreateConnection(ctx, connection)
			if err == nil {
				log.Infoln("New connection created with id:", id)
			}
		}

		if err != nil {
			log.Fatalln("could not create/modify connection", err)
		}
	}
	return err
}

func flattenSettings(settings map[string]string) string {
	result := ""
	for name, value := range settings {
		result += name + "=" + value + ";"
	}
	return result
}

func findExistingConnection(connections []*enigma.Connection, name string) string {
	for _, con := range connections {
		if con.Name == name {
			return con.Id
		}
	}
	return ""
}
