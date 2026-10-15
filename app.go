package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"querylight/internal/application"
)

// App struct
type App struct {
	service *application.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup composes the backend with the desktop lifetime context.
func (a *App) startup(ctx context.Context) {
	a.service = application.NewService(ctx)
}

func (a *App) shutdown(context.Context) { a.service.Shutdown() }
func (a *App) beforeClose(ctx context.Context) bool {
	if a.service == nil || !a.service.HasOpenWork() {
		return false
	}
	choice, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type: runtime.QuestionDialog, Title: "Close Querylight?",
		Message: "SQL tabs are kept in memory only. Closing discards their text, cancels running work, and closes database sessions. A running write may already have executed.",
		Buttons: []string{"Close Querylight", "Keep working"}, DefaultButton: "Keep working", CancelButton: "Keep working",
	})
	return err != nil || choice != "Close Querylight"
}
func (a *App) Connect(c application.ConnectionConfig) application.ConnectionReply {
	return a.service.Connect(c, false)
}
func (a *App) TestConnection(c application.ConnectionConfig) application.ConnectionReply {
	return a.service.Connect(c, true)
}
func (a *App) Disconnect(id string) application.ActionReply    { return a.service.Disconnect(id) }
func (a *App) ListSchemas(id string) application.MetadataReply { return a.service.ListSchemas(id) }
func (a *App) ListTables(id, schema string) application.MetadataReply {
	return a.service.ListTables(id, schema)
}
func (a *App) OpenEditor(id, schema string) application.ActionReply {
	return a.service.OpenEditor(id, schema)
}
func (a *App) CloseEditor(id string) application.ActionReply { return a.service.CloseEditor(id) }
func (a *App) StartExecution(req application.ExecutionRequest) application.ActionReply {
	return a.service.StartExecution(req)
}
func (a *App) StartInspection(id, schema, table string) application.ActionReply {
	return a.service.StartInspection(id, schema, table)
}
func (a *App) GetOperation(id string) application.OperationView { return a.service.GetOperation(id) }
func (a *App) CancelOperation(id string) application.ActionReply {
	return a.service.CancelOperation(id)
}
func (a *App) ReleaseOperation(id string) application.ActionReply {
	return a.service.ReleaseOperation(id)
}
