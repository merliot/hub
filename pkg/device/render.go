//go:build !tinygo

package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"
)

func (d *device) renderTmpl(w io.Writer, template string, data any) error {
	tmpl := d.templates.Lookup(template)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", template)
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		LogError("Rendering template", "err", err)
	}
	return err
}

func (d *device) renderSession(w io.Writer, template, sessionId string,
	level int, data map[string]any) error {
	data["sessionId"] = sessionId
	data["level"] = level
	return d.renderTmpl(w, template, data)
}

func (d *device) render(w io.Writer, sessionId, path, view string,
	level int, data map[string]any) error {

	path = strings.TrimPrefix(path, "/")
	template := path + "-" + view + ".tmpl"

	//LogDebug("render", "id", d.Id, "session-id", sessionId, "path", path,
	//	"view", view, "level", level, "template", template)
	if err := d.renderSession(w, template, sessionId, level, data); err != nil {
		return err
	}

	d.saveView(sessionId, view, level)

	return nil
}

func (d *device) renderPkt(w io.Writer, sessionId string, pkt *Packet) error {
	var data map[string]any

	view, level := d.lastView(sessionId)
	json.Unmarshal(pkt.Msg, &data)

	if data == nil {
		data = make(map[string]any)
	}

	//LogDebug("device.renderPkt", "id", d.Id, "view", view, "level", level, "pkt", pkt)
	return d.render(w, sessionId, pkt.Path, view, level, data)
}

func (p *Packet) render(w io.Writer, sessionId string) error {
	//LogDebug("Packet.render", "sessionId", sessionId, "pkt", p)

	s := p.server
	if s == nil {
		return fmt.Errorf("Packet.server not set")
	}

	d, exists := s.devices.get(p.Dst)
	if !exists {
		return fmt.Errorf("Invalid destination device id '%s'", p.Dst)
	}

	return d.renderPkt(w, sessionId, p)
}

func (d *device) renderTemplate(name string, data any) (template.HTML, error) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, name, data); err != nil {
		return template.HTML(""), err
	}
	return template.HTML(buf.String()), nil
}

func (d *device) renderView(sessionId, path, view string, level int) (template.HTML, error) {
	var buf bytes.Buffer

	if err := d.render(&buf, sessionId, path, view, level,
		map[string]any{}); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *device) renderChildrenWrite(w io.Writer, sessionId string, level int) error {
	return d.children.sortedByName(func(id string, child *device) error {
		view, _ := child.lastView(sessionId)
		if err := child.render(w, sessionId, "/device", view, level,
			map[string]any{}); err != nil {
			return err
		}
		return nil
	})
}

func (d *device) renderChildren(sessionId string, level int) (template.HTML, error) {
	var buf bytes.Buffer
	err := d.renderChildrenWrite(&buf, sessionId, level)
	return template.HTML(buf.String()), err
}
