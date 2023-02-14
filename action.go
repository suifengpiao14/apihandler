package onebehaviorentity

type ActionI interface {
	GetInput() []byte
	GetOutput() []byte
	GetLineSchemaInput() (lineschema string)
	GetLineSchemaOutput() (lineschema string)
	Do() (err error)
}

type Action struct {
	ActionI
	validateInput  *Validate
	validateOutput *Validate
}

func NewAction(actionI ActionI) (action *Action, err error) {
	var inputValidateI ValidateIFn = func() (lineschema string) {
		return actionI.GetLineSchemaInput()
	}
	validateInput, err := NewValidate(inputValidateI)
	if err != nil {
		return nil, err
	}

	var outputValidateI ValidateIFn = func() (lineschema string) {
		return actionI.GetLineSchemaOutput()
	}
	validateOutput, err := NewValidate(outputValidateI)
	if err != nil {
		return nil, err
	}

	action = &Action{ActionI: actionI, validateInput: validateInput, validateOutput: validateOutput}
	if err != nil {
		return nil, err
	}

	return
}

func (a Action) inputValidate() (err error) {
	input := a.GetInput()
	inputStr := string(input)
	err = a.validateInput.Validate(inputStr)
	if err != nil {
		return err
	}
	return nil
}
func (a Action) outputValidate() (err error) {
	output := a.GetInput()
	outputStr := string(output)
	err = a.validateOutput.Validate(outputStr)
	if err != nil {
		return err
	}
	return nil
}

func (a Action) Do() (err error) {
	err = a.inputValidate()
	if err != nil {
		return err
	}
	err = a.ActionI.Do()
	if err != nil {
		return err
	}
	err = a.outputValidate()
	if err != nil {
		return err
	}
	return
}
