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
	choice, err := runtime.MessageDialog(ctx, closeDialogOptions())
	return preventClose(choice, err)
}

func closeDialogOptions() runtime.MessageDialogOptions {
	// Wails' Windows QuestionDialog uses native Yes/No buttons and returns
	// those canonical strings; custom button captions are ignored there.
	return runtime.MessageDialogOptions{
		Type: runtime.QuestionDialog, Title: "Close Querylight?",
		Message: "Close Querylight and discard in-memory SQL text and local table drafts? Closing cancels running work and closes database sessions. A running SQL write may already have executed. Choose Yes to close or No to keep working.",
		Buttons: []string{"Yes", "No"}, DefaultButton: "No", CancelButton: "No",
	}
}

func preventClose(choice string, err error) bool {
	return err != nil || choice != "Yes"
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
