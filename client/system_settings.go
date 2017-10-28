package client

import (
	"encoding/json"
	"fmt"
)

type SystemSettings struct {
	userStreamAcl   *StreamAcl
	systemStreamAcl *StreamAcl
}

func NewSystemSettings(
	userStreamAcl *StreamAcl,
	systemStreamAcl *StreamAcl,
) *SystemSettings {
	return &SystemSettings{userStreamAcl, systemStreamAcl}
}

func (s *SystemSettings) UserStreamAcl() *StreamAcl { return s.userStreamAcl }

func (s *SystemSettings) SystemStreamAcl() *StreamAcl { return s.systemStreamAcl }

func (s *SystemSettings) String() string {
	return fmt.Sprintf("&{userStreamAcl:%+v systemStreamAcl:%+v}", s.userStreamAcl, s.systemStreamAcl)
}

type systemSettingsJson struct {
	UserStreamAcl   *StreamAcl `json:"$userStreamAcl,omitempty"`
	SystemStreamAcl *StreamAcl `json:"$systemStreamAcl,omitempty"`
}

func (s *SystemSettings) MarshalJSON() ([]byte, error) {
	return json.Marshal(systemSettingsJson{
		UserStreamAcl:   s.userStreamAcl,
		SystemStreamAcl: s.systemStreamAcl,
	})
}

func (s *SystemSettings) UnmarshalJSON(data []byte) error {
	ss := systemSettingsJson{}
	if err := json.Unmarshal(data, &ss); err != nil {
		return err
	}
	s.userStreamAcl = ss.UserStreamAcl
	s.systemStreamAcl = ss.SystemStreamAcl
	return nil
}

func SystemSettingsFromJsonBytes(data []byte) (*SystemSettings, error) {
	systemSettings := &SystemSettings{}
	return systemSettings, json.Unmarshal(data, systemSettings)
}
