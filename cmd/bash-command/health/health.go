package health

import (
	"github.com/Qihoo360/doraemon/cmd/bash-command/models"
)

type EnsureDatabase struct {
}

func (e *EnsureDatabase) Health() error {
	_, err := models.Ormer().QueryTable(new(models.User)).Count()
	if err != nil {
		return err
	}
	return nil
}
