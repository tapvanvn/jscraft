package tokenize

//Meaning inteface for language meaning process
type Meaning interface {
	GetNextMeaningToken() *BaseToken
}
