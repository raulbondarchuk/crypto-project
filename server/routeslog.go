package server

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

func LogRoutes(r *gin.Engine) {
	type row struct{ Method, Path, Handler string }

	group := map[string][]row{}
	for _, rt := range r.Routes() {
		key := first(rt.Path)
		group[key] = append(group[key], row{rt.Method, rt.Path, short(rt.Handler)})
	}
	gKeys := make([]string, 0, len(group))
	for k := range group {
		gKeys = append(gKeys, k)
	}
	sort.Strings(gKeys)

	rank := map[string]int{"GET": 1, "POST": 2, "PUT": 3, "PATCH": 4, "DELETE": 5}

	const (
		wM = 6
		wP = 60
		wH = 36
	)
	sep := " │ "
	lineW := 1 + wM + len(sep) + wP + len(sep) + wH

	borderH := "├" + strings.Repeat("─", lineW-1)
	top := "┌" + strings.Repeat("─", lineW-1)
	bottom := "└" + strings.Repeat("─", lineW-1)

	var out strings.Builder
	out.WriteString(fmt.Sprintf("\nRegistered routes (%d):\n", len(r.Routes())))
	out.WriteString(top + "\n")

	mCol := crop("METHOD", wM)
	pCol := crop("PATH", wP)
	hCol := crop("HANDLER", wH)
	out.WriteString(fmt.Sprintf("│%-*s%s%-*s%s%-*s\n", wM, mCol, sep, wP, pCol, sep, wH, hCol))
	out.WriteString(borderH + "\n")

	for gi, g := range gKeys {
		rows := group[g]
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Path == rows[j].Path {
				return rank[rows[i].Method] < rank[rows[j].Method]
			}
			return rows[i].Path < rows[j].Path
		})
		for _, rw := range rows {
			mCol := crop(rw.Method, wM)
			pCol := crop(rw.Path, wP)
			hCol := crop(rw.Handler, wH)
			mCol = color(rw.Method) + fmt.Sprintf("%-*s", wM, mCol) + reset

			out.WriteString(fmt.Sprintf("│%s%s%-*s%s%-*s\n",
				mCol, sep, wP, pCol, sep, wH, hCol))
		}
		if gi < len(gKeys)-1 {
			out.WriteString(borderH + "\n")
		}
	}
	out.WriteString(bottom + "\n")
	out.WriteString(reset)

	fmt.Print(out.String())
}

/* utils */

func crop(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w <= 1 {
		return string(r[:w])
	}
	return string(r[:w-1]) + "…"
}
func short(fq string) string {
	if i := strings.LastIndex(fq, "."); i != -1 {
		return fq[i+1:]
	}
	return fq
}
func first(p string) string {
	if p == "/" {
		return "/"
	}
	return strings.Split(strings.TrimPrefix(p, "/"), "/")[0]
}

/* ANSI colours (как у gin.Logger) */

const (
	green  = "\033[32m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	red    = "\033[31m"
	blue   = "\033[34m"
	mag    = "\033[35m"
	reset  = "\033[0m"
)

func color(m string) string {
	switch m {
	case "GET":
		return green
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return blue
	case "HEAD":
		return mag
	default:
		return ""
	}
}
