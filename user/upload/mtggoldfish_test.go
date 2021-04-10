package upload

import "testing"

func TestHandleImportForMTGGoldfish(t *testing.T) {
	content := `Card,Set ID,Set Name,Quantity,Foil
	"Forest",DOM,"",1,""
	"Island",DOM,"",1,""
	"Mountain",DOM,"",1,""
	"Plains",DOM,"",1,""
	"Swamp",DOM,"",1,""`
	testHandleImport(t, content)
}
