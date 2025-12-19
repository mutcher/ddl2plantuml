package oracle

import "github.com/icpd/ddl2plantuml/driver/common"

// SqlStateHandlerImpl

type SqlStateHandlerImpl struct {
	Tables       *common.Tables
	CurrentState MutableState
}

// GetTables implements SqlStateHandler.
func (s *SqlStateHandlerImpl) GetTables() *common.Tables {
	return s.CurrentState.GetTables()
}

func (s *SqlStateHandlerImpl) Reset() {
	s.Tables = &common.Tables{}
	s.CurrentState = &NoneState{s.Tables}
}

func (s *SqlStateHandlerImpl) InjectWord(word string) error {
	state, err := s.CurrentState.InjectWord(word)
	if err != nil {
		return err
	}

	s.CurrentState = state
	return nil
}
