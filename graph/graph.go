package graph

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"yoyuuhi/kolink/request"
	"yoyuuhi/kolink/state"

	"github.com/awalterschulze/gographviz"
	"github.com/goccy/go-graphviz"
	"github.com/juju/errors"
)

func DrawGraphs(callerFileStateMap map[string][]state.FuncState, requestDef request.RequestDef) error {
	calleeFileStateMap := map[string][]state.FuncState{}
	for k, vs := range callerFileStateMap {
		for _, v := range vs {
			callee := v.File
			if _, e := calleeFileStateMap[callee]; !e {
				calleeFileStateMap[callee] = []state.FuncState{}
			}

			caller := strings.Split(k, ".")[0]
			calleeFileStateMap[callee] = append(calleeFileStateMap[callee], state.FuncState{
				File:     caller,
				Function: v.Function,
			})
		}
	}

	for k, vs := range calleeFileStateMap {
		if val, e := requestDef.FileParamMap[k]; e && val.Split {
			folderName := strings.Split(k, ".")[0]
			path := filepath.Join(".", requestDef.OutDir, folderName)

			groupedFS := map[string][]state.FuncState{}
			for _, v := range vs {
				if _, e := groupedFS[v.Function]; !e {
					groupedFS[v.Function] = []state.FuncState{}
				}
				groupedFS[v.Function] = append(groupedFS[v.Function], v)
			}
			for fileName, nvs := range groupedFS {
				if err := drawGraph(requestDef, fileName, k, nvs, path); err != nil {
					return errors.Trace(err)
				}
			}
		} else {
			path := filepath.Join(".", requestDef.OutDir)
			fileName := strings.Split(k, ".")[0]
			if err := drawGraph(requestDef, fileName, k, vs, path); err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}

func drawGraph(requestDef request.RequestDef, fileName string, k string, vs []state.FuncState, path string) error {
	g := gographviz.NewGraph()
	if err := g.SetName("g"); err != nil {
		return errors.Trace(err)
	}
	for _, attribute := range requestDef.FileParamMap[k].Attributes {
		g.AddAttr("g", attribute.Name, attribute.Value)
	}

	if err := g.AddSubGraph("g", "cluster_caller", map[string]string{
		"style":     "filled",
		"fillcolor": "lightgrey",
		"rank":      "same",
		"label":     "caller",
	}); err != nil {
		return err
	}
	if err := g.AddSubGraph("g", "cluster_callee", map[string]string{
		"fillcolor": "blue",
		"rank":      "same",
		"label":     "callee",
	}); err != nil {
		return errors.Trace(err)
	}

	drawn := map[string]map[string]bool{}
	for _, v := range vs {
		caller := v.File
		f := v.Function

		if err := g.AddNode("cluster_caller", caller, map[string]string{
			"style":     "wedged",
			"fillcolor": "blue",
		}); err != nil {
			return errors.Trace(err)
		}
		if err := g.AddNode("cluster_callee", f, map[string]string{
			"style": "rounded",
		}); err != nil {
			return errors.Trace(err)
		}

		if _, e := drawn[caller]; !e {
			drawn[caller] = map[string]bool{}
		}
		if drawn[caller][f] {
			continue
		}
		drawn[caller][f] = true
		if err := g.AddEdge(caller, f, false, nil); err != nil {
			return errors.Trace(err)
		}
	}

	gg := graphviz.New()
	graph, err := graphviz.ParseBytes([]byte(g.String()))
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		gg.Close()
	}()

	if v, e := requestDef.FileParamMap[k]; e && v.Layout != "" {
		graph.SetLayout(v.Layout)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Trace(err)
	}
	if err := os.MkdirAll(absPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	if err := gg.RenderFilename(graph, graphviz.PNG, absPath+"/"+fileName+".png"); err != nil {
		log.Fatal(err)
	}

	return nil
}
