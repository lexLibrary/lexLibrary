// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestDocument(t *testing.T) {
	var user *app.User
	reset := func(t *testing.T) {
		t.Helper()

		user = resetAdmin(t, "admin", "adminpassword").User()
	}

	t.Run("New Document", func(t *testing.T) {
		reset(t)

		draft, err := user.NewDocument()
		if err != nil {
			t.Fatalf("Error adding new document: %s", err)
		}

		if draft == nil {
			t.Fatalf("Draft document was nil")
		}

		if draft.Version != 0 {
			t.Fatalf("Draft version is incorrect.  Expected %d, got %d", 0, draft.Version)
		}

		if draft.ID.IsNil() {
			t.Fatalf("Draft ID is nil")
		}

		if !draft.DocumentContent.ID.IsNil() {
			t.Fatalf("Document ID is not nil")
		}

		assertRow(t, data.NewQuery(`
			select count(*) from document_drafts where id = {{arg "id"}}
		`).QueryRow(data.Arg("id", draft.ID)), 1)

	})

	t.Run("New Draft", func(t *testing.T) {
		reset(t)
		draft, err := user.NewDocument()
		if err != nil {
			t.Fatalf("Error adding new draft: %s", err)
		}

		t.Run("Save", func(t *testing.T) {
			newTitle := "new Title"
			newContent := "<h2>New Content</h2>"
			newTags := []string{"newtag1", "newtag2", "newtag3", "newtag1", "newtag1", "newtag2"}

			d := *draft

			err = d.Save("", newContent, newTags, draft.Version)
			if !app.IsFail(err) {
				t.Fatalf("No failure on empty title: %s", err)
			}

			d = *draft
			err = d.Save(newTitle, newContent, newTags, 3)
			if !app.IsFail(err) {
				t.Fatalf("No failure on incorrect version: %s", err)
			}

			d = *draft
			err = d.Save(newTitle, newContent, []string{fmt.Sprintf("Long %70s", "tag value")}, 3)
			if !app.IsFail(err) {
				t.Fatalf("No failure on incorrect version: %s", err)
			}

			d = *draft
			err = d.Save(newTitle, newContent, newTags, draft.Version)
			if err != nil {
				t.Fatalf("Error Saving draft: %s", err)
			}

			if d.Version != 1 {
				t.Fatalf("Incorrect draft version. Expected %d, got %d", 1, d.Version)
			}

			if d.Title != newTitle {
				t.Fatalf("Incorrect title. Expected %s got %s", newTitle, d.Title)
			}

			if d.Content != newContent {
				t.Fatalf("Incorrect content. Expected %s got %s", newContent, d.Content)
			}

			for i := range newTags {
				found := false
				for j := range d.Tags {
					if newTags[i] == d.Tags[j].Value {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("Tag %s not found in draft tags", newTags[i])
				}
			}

			if len(d.Tags) != 3 {
				t.Fatalf("Draft contains duplicate tags.  Expected %d, got %d", 3, len(d.Tags))
			}

		})

		t.Run("Publish New", func(t *testing.T) {
			doc, err := draft.Publish()
			if err != nil {
				t.Fatalf("Error publishing draft: %s", err)
			}

			if doc == nil {
				t.Fatalf("Published Document is nil")
			}

			if doc.Title != draft.Title {
				t.Fatalf("Published doc title doesn't match draft. Expected %s, got %s", draft.Title,
					doc.Title)
			}

			if doc.Content != draft.Content {
				t.Fatalf("Published doc content doesn't match draft. Expected %s, got %s", draft.Content,
					doc.Content)
			}

			if doc.DraftID != draft.ID {
				t.Fatalf("Published doc has invalid draft id. Expected %s, got %s", draft.ID,
					doc.DraftID)
			}

			if doc.ID.IsNil() {
				t.Fatalf("Doc has nil ID")
			}

			for i := range draft.Tags {
				found := false
				for j := range doc.Tags {
					if draft.Tags[i].Value == doc.Tags[j].Value &&
						draft.Tags[i].Type == doc.Tags[j].Type {
						found = true
					}
				}

				if !found {
					t.Fatalf("Draft tag %s not found in new Doc: ", draft.Tags[i].Value)
				}
			}

			assertRow(t, data.NewQuery(`
				select id, version, draft_id, title, content from documents where id = {{arg "id"}}
			`).QueryRow(data.Arg("id", doc.ID)),
				doc.ID, doc.Version, draft.ID, draft.Title, draft.Content)
			assertRow(t, data.NewQuery(`
				select count(*) from document_tags where document_id = {{arg "id"}}
			`).QueryRow(data.Arg("id", doc.ID)), 0)

			assertRow(t, data.NewQuery(`
				select count(*) from document_drafts where id = {{arg "id"}}
			`).QueryRow(data.Arg("id", draft.ID)), 0)

			assertRow(t, data.NewQuery(`
				select count(*) from document_draft_tags where draft_id = {{arg "id"}}
			`).QueryRow(data.Arg("id", draft.ID)), 0)

			assertRow(t, data.NewQuery(`
				select count(*) from document_history 
				where draft_id = {{arg "draft_id"}} and document_id = {{arg "document_id"}}
			`).QueryRow(data.Arg("draft_id", draft.ID), data.Arg("document_id", doc.ID)), 0)
		})

	})

	t.Run("Existing Document Draft", func(t *testing.T) {
	})

}
