package models

// StoredProcedureCall estructura para almacenar la llamada al procedimiento almacenado
type StoredProcedureCall struct {
	ProcedureName string        `json:"ProcedureName"`
	Inputs        []interface{} `json:"Inputs"`
	Outputs       int           `json:"Outputs"`
}
