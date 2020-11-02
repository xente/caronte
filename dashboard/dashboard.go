package dashboard

import (
	"Caronte/core"
	"Caronte/orchestrator/discovery"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"go.uber.org/zap"
	"html/template"
	"net/http"
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
