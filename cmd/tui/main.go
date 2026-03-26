package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"partialupdate/internal/models"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ---------------------------------------------------------------------
// HTTP helper – a tiny wrapper around http.Client that returns body as string
// ---------------------------------------------------------------------

var httpClient = &http.Client{}

func doRequest(method, url string, payload any) (int, string, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("json marshal: %w", err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("new request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody), nil
}

// ---------------------------------------------------------------------
// UI helpers
// ---------------------------------------------------------------------

// showResult creates a modal with the HTTP status and body.
func showResult(app *tview.Application, title string, status int, body string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[yellow]Status:[-] %d\n\n[white]%s[-]", status, body)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(rootFlex, true).SetFocus(menu)
		})
	modal.SetTitle(title).SetBorder(true)
	app.SetRoot(modal, false).SetFocus(modal)
}

// ---------------------------------------------------------------------
// UI components
// ---------------------------------------------------------------------

var (
	app      *tview.Application
	menu     *tview.List
	rootFlex *tview.Flex
)

func main() {
	app = tview.NewApplication()

	// -------------------------- MENU --------------------------
	menu = tview.NewList().
		AddItem("🟢 Create", "POST /v1/person", 'c', showCreateForm).
		AddItem("📄 List", "GET /v1/person", 'l', listPersons).
		AddItem("🔎 Get", "GET /v1/person/{id}", 'g', showGetForm).
		AddItem("✏️ Full Update", "PUT /v1/person/{id}", 'u', showFullUpdateForm).
		AddItem("🩹 Partial Update", "PATCH /v1/person/{id}", 'p', showPatchForm).
		AddItem("❌ Delete", "DELETE /v1/person/{id}", 'd', showDeleteForm).
		AddItem("🚪 Exit", "Quit the program", 'q', func() { app.Stop() })
		// SetBorder(true).SetTitle("Menu – choose an action")
	// SetShortcutColor(tcell.ColorGreen)

	// -------------------------- MAIN LAYOUT --------------------------
	rootFlex = tview.NewFlex().
		AddItem(menu, 30, 1, true) // left pane (fixed width)
	// right side will be filled by forms / tables when needed

	if err := app.SetRoot(rootFlex, true).Run(); err != nil {
		panic(err)
	}
}

// ---------------------------------------------------------------------
// CREATE
// ---------------------------------------------------------------------

func showCreateForm() {
	form := tview.NewForm()
	form.AddInputField("ID", "", 20, nil, nil).
		AddInputField("Name", "", 30, nil, nil).
		AddCheckbox("Alive", false, nil).
		AddInputField("Age", "", 5, func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			return err == nil
		}, nil).
		AddButton("Create", func() {
			id := form.GetFormItemByLabel("ID").(*tview.InputField).GetText()
			name := form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
			alive := form.GetFormItemByLabel("Alive").(*tview.Checkbox).IsChecked()
			ageStr := form.GetFormItemByLabel("Age").(*tview.InputField).GetText()

			age, err := strconv.Atoi(ageStr)
			if err != nil {
				showResult(app, "Validation error", http.StatusBadRequest,
					"`Age` must be a number")
				return
			}

			// ---- Build the payload -------------------------------------------------
			person := models.Person{
				ID:    id,
				Name:  name,
				Alive: alive,
				Age:   age,
			}

			status, body, err := doRequest(
				http.MethodPost,
				"http://localhost:8080/v1/person",
				&person,
			)
			if err != nil {
				showResult(app, "Create – error", status, err.Error())
				return
			}
			showResult(app, "Create – response", status, body)
		}).
		AddButton("Cancel", func() { app.SetRoot(rootFlex, true).SetFocus(menu) })
	form.SetBorder(true).SetTitle("Create a new Person").SetTitleAlign(tview.AlignLeft)

	// replace the right pane with the form
	// rootFlex.AddItem(form, 0, 1, true)
	replaceRight(form)
	app.SetFocus(form)
}

// ---------------------------------------------------------------------
// LIST
// ---------------------------------------------------------------------

func listPersons() {
	status, body, err := doRequest(http.MethodGet, "http://localhost:8080/v1/person", nil)
	if err != nil {
		showResult(app, "List – error", 0, err.Error())
		return
	}
	if status != http.StatusOK {
		showResult(app, "List – non‑200", status, body)
		return
	}

	// Parse JSON array into []Person so we can show a nice table
	var people []models.Person
	if err := json.Unmarshal([]byte(body), &people); err != nil {
		showResult(app, "List – bad JSON", status, body)
		return
	}

	table := tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false)

	// Header
	headers := []string{"ID", "Name", "Alive", "Age"}
	for c, h := range headers {
		table.SetCell(0, c,
			tview.NewTableCell("[::b]"+h).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter))
	}
	// Rows
	for r, p := range people {
		table.SetCell(r+1, 0,
			tview.NewTableCell(p.ID).SetTextColor(tcell.ColorWhite))
		table.SetCell(r+1, 1,
			tview.NewTableCell(p.Name).SetTextColor(tcell.ColorWhite))
		table.SetCell(r+1, 2,
			tview.NewTableCell(strconv.FormatBool(p.Alive)).SetTextColor(tcell.ColorWhite))
		table.SetCell(r+1, 3,
			tview.NewTableCell(strconv.Itoa(p.Age)).SetTextColor(tcell.ColorWhite))
	}

	table.SetBorder(true).SetTitle("People (press Esc to go back)")

	// When the user hits Esc, go back to the menu
	table.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.SetRoot(rootFlex, true).SetFocus(menu)
		}
	})

	// Replace the right pane with the table
	// rootFlex.AddItem(table, 0, 2, true)
	replaceRight(table)
	app.SetFocus(table)
}

// ---------------------------------------------------------------------
// GET ONE
// ---------------------------------------------------------------------

func showGetForm() {
	form := tview.NewForm()
	form.AddInputField("ID", "", 20, nil, nil).
		AddButton("Get", func() {
			id := form.GetFormItemByLabel("ID").(*tview.InputField).GetText()
			url := fmt.Sprintf("http://localhost:8080/v1/person/%s", id)
			status, body, err := doRequest(http.MethodGet, url, nil)
			if err != nil {
				showResult(app, "Get – error", 0, err.Error())
				return
			}
			showResult(app, fmt.Sprintf("GET /v1/person/%s", id), status, body)
		}).
		AddButton("Cancel", func() { app.SetRoot(rootFlex, true).SetFocus(menu) })
	form.SetBorder(true).SetTitle("Get a Person by ID")
	// rootFlex.AddItem(form, 0, 1, true)
	replaceRight(form)
	app.SetFocus(form)
}

// ---------------------------------------------------------------------
// FULL UPDATE (PUT)
// ---------------------------------------------------------------------

func showFullUpdateForm() {
	form := tview.NewForm()
	form.AddInputField("ID (path param)", "", 20, nil, nil).
		AddInputField("Name", "", 30, nil, nil).
		AddCheckbox("Alive", false, nil).
		AddInputField("Age", "", 5, func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			return err == nil
		}, nil).
		AddButton("Update", func() {
			id := form.GetFormItemByLabel("ID").(*tview.InputField).GetText()
			name := form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
			alive := form.GetFormItemByLabel("Alive").(*tview.Checkbox).IsChecked()
			ageStr := form.GetFormItemByLabel("Age").(*tview.InputField).GetText()
			age, _ := strconv.Atoi(ageStr)

			person := models.Person{Name: name, Alive: alive, Age: age}
			url := fmt.Sprintf("http://localhost:8080/v1/person/%s", id)
			status, body, err := doRequest(http.MethodPut, url, person)
			if err != nil {
				showResult(app, "Full Update – error", 0, err.Error())
				return
			}
			showResult(app, fmt.Sprintf("PUT /v1/person/%s", id), status, body)
		}).
		AddButton("Cancel", func() { app.SetRoot(rootFlex, true).SetFocus(menu) })
	form.SetBorder(true).SetTitle("Full Update (PUT)")
	// rootFlex.AddItem(form, 0, 1, true)
	replaceRight(form)
	app.SetFocus(form)
}

// ---------------------------------------------------------------------
// PARTIAL UPDATE (PATCH)
// ---------------------------------------------------------------------

func showPatchForm() {
	form := tview.NewForm()
	form.AddInputField("ID", "", 20, nil, nil).
		AddInputField("Name (optional)", "", 30, nil, nil).
		AddCheckbox("Alive (optional)", false, nil).
		AddInputField("Age (optional)", "", 5, func(textToCheck string, lastChar rune) bool {
			if textToCheck == "" {
				return true // allow empty field → means “not set”
			}
			_, err := strconv.Atoi(textToCheck)
			return err == nil
		}, nil).
		AddButton("Patch", func() {
			id := form.GetFormItemByLabel("ID (path param)").(*tview.InputField).GetText()
			name := form.GetFormItemByLabel("Name (optional)").(*tview.InputField).GetText()
			aliveChecked := form.GetFormItemByLabel("Alive (optional)").(*tview.Checkbox).IsChecked()
			ageStr := form.GetFormItemByLabel("Age (optional)").(*tview.InputField).GetText()

			// Build a partial payload – omit zero‑value fields so the server can ignore them.
			patch := models.Person{}
			if name != "" {
				patch.Name = name
			}
			if ageStr != "" {
				age, _ := strconv.Atoi(ageStr)
				patch.Age = age
			}
			// We can't differentiate “checkbox left unchecked because I don’t want to change” from “I want to set false”.
			// For this demo we always send the boolean value; real life would need a tri‑state (nil/true/false) design.
			patch.Alive = aliveChecked

			url := fmt.Sprintf("http://localhost:8080/v1/person/%s", id)
			status, body, err := doRequest(http.MethodPatch, url, patch)
			if err != nil {
				showResult(app, "Patch – error", 0, err.Error())
				return
			}
			showResult(app, fmt.Sprintf("PATCH /v1/person/%s", id), status, body)
		}).
		AddButton("Cancel", func() { app.SetRoot(rootFlex, true).SetFocus(menu) })
	form.SetBorder(true).SetTitle("Partial Update (PATCH)")
	// rootFlex.AddItem(form, 0, 1, true)
	replaceRight(form)
	app.SetFocus(form)
}

// ---------------------------------------------------------------------
// DELETE
// ---------------------------------------------------------------------

func showDeleteForm() {
	form := tview.NewForm()
	form.AddInputField("ID", "", 20, nil, nil).
		AddButton("Delete", func() {
			id := form.GetFormItemByLabel("ID").(*tview.InputField).GetText()
			url := fmt.Sprintf("http://localhost:8080/v1/person/%s", id)
			status, body, err := doRequest(http.MethodDelete, url, nil)
			if err != nil {
				showResult(app, "Delete – error", 0, err.Error())
				return
			}
			showResult(app, fmt.Sprintf("DELETE /v1/person/%s", id), status, body)
		}).
		AddButton("Cancel", func() { app.SetRoot(rootFlex, true).SetFocus(menu) })
	form.SetBorder(true).SetTitle("Delete a Person")
	// rootFlex.AddItem(form, 0, 1, true)
	replaceRight(form)
	app.SetFocus(form)
}

// replaceRight replaces whatever is currently displayed on the right side of
// the root Flex (index 1) with the supplied primitive.
// If there is no right‑hand item yet, it simply adds one.
func replaceRight(p tview.Primitive) {
	// The right pane is always at index 1 (menu is index 0).
	// Remove the old pane if it exists.
	if rootFlex.GetItemCount() > 1 {
		// GetItem returns (item, proportion, focus) – we ignore the latter.
		old := rootFlex.GetItem(1)
		if old != nil {
			rootFlex.RemoveItem(old)
		}
	}
	// Add the new pane, giving it a flexible proportion (1) and focusability.
	rootFlex.AddItem(p, 0, 2, true)
}

// helper that converts the selected string into a *bool
func parseTriState(selection string) *bool {
	switch selection {
	case "true":
		v := true
		return &v
	case "false":
		v := false
		return &v
	default: // "— unchanged —" or any other value
		return nil
	}
}
