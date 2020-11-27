package dashboard

import (
	"Caronte/core"
	"Caronte/orchestrator/discovery"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"go.uber.org/zap"
)

type PageData struct {
	Services map[string]core.CaronteService
}

func Dashboard(port int) {
	templatesBox := packr.New("Templates", ".")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Services: discovery.GetActiveServices(),
		}
		templateLayout, err := templatesBox.FindString("dashboard.html")
		if err != nil {
			zap.S().Error(err)
		}
		t := template.New("")
		t.Parse(templateLayout)
		t.Execute(w, data)
	})

	http.ListenAndServe(fmt.Sprint(":", port), nil)
}
