package college

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Institution struct {
	DisplayName            string `json:"displayName"`
	PrimaryKey             string `json:"primaryKey"`
	AliasName              string `json:"alias"`
	SchoolType             string `json:"schoolType"`
	UrlName                string `json:"urlName"`
	Region                 string `json:"region"`
	Location               string `json:"location"`
	State                  string `json:"state"`
	City                   string `json:"city"`
	Zip                    string `json:"zipCode"`
	RankingDisplayName     string `json:"rankingDisplayName"`
	RankingDisplayRank     string `json:"rankingDisplayRank"`
	RankingDisplayScore    string `json:"rankingDisplayScore"`
	RankingFullDisplayText string `json:"rankingFullDisplayText"`
	InstitutionalControl   string `json:"institutionalControl"`
	PrimaryPhotoThumb      string `json:"primaryPhotoThumb"`
}

type CommonFields struct {
	FieldName    string `json:"fieldName"`
	DisplayValue string `json:"displayValue"`
	//RawValue     float32 `json:"rawValue"`
}

type Tuition struct {
	RawValue     int         `json:"rawValue"`
	FieldType    string      `json:"fieldType"`
	DisplayValue interface{} `json:"displayValue"`
	FieldName    string      `json:"fieldName"`
}

func decodeNameValue(data any) string {
	switch v := data.(type) {
	case string:
		return v
	case []interface{}:
		var DisplayValueStrings []string
		for _, vv := range data.([]interface{}) {
			kv := vv.(map[string]interface{})
			DisplayValueStrings = append(DisplayValueStrings, fmt.Sprintf("%s: %s", kv["name"].(string), kv["value"].(string)))
		}
		return strings.Join(DisplayValueStrings, "%")
	}
	return ""
}

func (t *Tuition) UnmarshalJSON(data []byte) error {
	type Alias Tuition
	aux := &struct {
		*Alias
		DisplayValueJson any `json:"displayValue"`
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.DisplayValue = decodeNameValue(aux.DisplayValueJson)
	return nil
}

type Enrollment struct {
	DataQaId     string      `json:"dataQaId"`
	RawValue     int         `json:"rawValue"`
	FieldName    string      `json:"totalUndergradEnrollment"`
	DisplayValue interface{} `json:"displayValue"`
}

func (t *Enrollment) UnmarshalJSON(data []byte) error {
	type Alias Enrollment
	aux := &struct {
		*Alias
		DisplayValueJson any `json:"displayValue"`
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.DisplayValue = decodeNameValue(aux.DisplayValueJson)

	return nil
}

type CostAfterAid struct {
	CommonFields
}

type PercentReceivingAid struct {
	CommonFields
}

type AcceptanceRate struct {
	CommonFields
}
type HsGpaAvg struct {
	CommonFields
}
type SATAvg struct {
	CommonFields
}

type ACTAvg struct {
	CommonFields
}

type EngineeringRepScore struct {
	CommonFields
}
type BusinessRepScore struct {
	CommonFields
}
type ComputerScienceRepScore struct {
	CommonFields
}
type NursingRepScore struct {
	CommonFields
}
type EducationRepScore struct {
	CommonFields
}
type LawRepScore struct {
	CommonFields
}
type MedicineRepScore struct {
	CommonFields
}
type PsychologyRepScore struct {
	CommonFields
}
type CriminalJusticeRepScore struct {
	CommonFields
}
type CommunicationsRepScore struct {
	CommonFields
}
type BiologyRepScore struct {
	CommonFields
}
type EnglishRepScore struct {
	CommonFields
}
type TestAvgs struct {
	DisplayValue []map[string]string `json:"displayValue"`
}

type SearchData struct {
	Tuition             Tuition             `json:"tuition"`
	Enrollment          Enrollment          `json:"enrollment"`
	CostAfterAid        CostAfterAid        `json:"costAfterAid"`
	PercentReceivingAid PercentReceivingAid `json:"percentReceivingAid"`
	AcceptanceRate      AcceptanceRate      `json:"acceptanceRate"`
	HsGpaAvg            HsGpaAvg            `json:"hsGpaAvg"`
	SATAvg              SATAvg              `json:"satAvg"`
	ACTAvg              ACTAvg              `json:"actAvg"`
	EngineeringRepScore EngineeringRepScore `json:"engineeringRepScore"`
	BusinessRepScore    BusinessRepScore    `json:"businessRepScore"`
	//ComputerScienceRepScore ComputerScienceRepScore `json:"computerScienceRepScore"`
	NursingRepScore         NursingRepScore         `json:"nursingRepScore"`
	EducationRepScore       EducationRepScore       `json:"educationRepScore"`
	LawRepScore             LawRepScore             `json:"lawRepScore"`
	MedicineRepScore        MedicineRepScore        `json:"medicineRepScore"`
	PsychologyRepScore      PsychologyRepScore      `json:"psychologyRepScore"`
	CriminalJusticeRepScore CriminalJusticeRepScore `json:"criminalJusticeRepScore"`
	CommunicationsRepScore  CommunicationsRepScore  `json:"communicationsRepScore"`
	BiologyRepScore         BiologyRepScore         `json:"biologyRepScore"`
	EnglishRepScore         EnglishRepScore         `json:"englishRepScore"`
	TestAvgs                TestAvgs                `json:"testAvgs"`
}

type College struct {
	Institution Institution `json:"institution"`
	SearchData  SearchData  `json:"searchData"`
	Blurb       string      `json:"blurb"`
}

type Data struct {
	Items  []College `json:"items"`
	Sort   string    `json:"sort"`
	SortBy string    `json:"sortBy"`
}

type SchoolData struct {
	Data Data `json:"data"`
}
