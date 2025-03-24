package userutils

import (
    "encoding/json"
    "fmt"
)

type ApplicationStatus string 

const (
    StatusApplied      ApplicationStatus = "Applied"
    StatusScreen       ApplicationStatus = "Screen"
    StatusInterviewing ApplicationStatus = "Interviewing"
    StatusOffer        ApplicationStatus = "Offer"
    StatusRejected     ApplicationStatus = "Rejected"
    StatusGhosted      ApplicationStatus = "Ghosted"
)

func (s ApplicationStatus) IsValid() bool {
    switch s {
    case StatusApplied, StatusScreen, StatusInterviewing, StatusOffer, StatusRejected, StatusGhosted:
        return true
    }
    return false
}

// json.Unmarshaler interface implementation sees that Status fieldd is of type ApplicationStatus
// then checks if ApplicationStatus implements the json.Unmarshaler interface.
// Since it does (this is the implementation), it calls the UnmarshhalJSON method automatically for each
// field of this type. So, all we have to do is make every status as type ApplicationStatus
// and during JSON unmarshalling this methohd will be called
func (s *ApplicationStatus) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    
    status := ApplicationStatus(str)
    if !status.IsValid() {
        return fmt.Errorf("invalid status: %s", str)
    }
    
    *s = status
    return nil
}