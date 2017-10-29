package contentstream

import (
	"testing"
)

func TestOperandTJSpacing(t *testing.T) {

	content := `BT
	[(are)-328(h)5(ypothesized)-328(to)-327(in\003uence)-328(the)-328(stability)-328(of)-328(the)-328(upstream)-327(glaciers,)-328(and)-328(thus)-328(of)-328(the)-328(entire)-327(ice)-328(sheet)]TJ
	ET`
	referenceText := "are hypothesized to in\003uence the stability of the upstream glaciers, and thus of the entire ice sheet"

	cStreamParser := NewContentStreamParser(content)

	text, err := cStreamParser.ExtractText()
	if err != nil {
		t.Error()
	}

	if text != referenceText {
		t.Fail()
	}

}
