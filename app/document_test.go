// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
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
		title := "test title"
		content := "<div>TestContent</div>"
		tags := []string{"tag1", "tag2", "tag3", "tag1", "tag1", "tag2"}

		_, err := user.NewDocument("", content, tags)
		if !app.IsFail(err) {
			t.Fatalf("Adding document without title did not fail: %s", err)
		}

		draft, err := user.NewDocument(title, content, tags)
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

		if draft.Title != title {
			t.Fatalf("Draft tite is incorrect. Expected %s, got %s", title, draft.Title)
		}

		if draft.Content != content {
			t.Fatalf("Draft content is incorrect. Expected %s, got %s", content, draft.Content)
		}

		for i := range tags {
			found := false
			for j := range draft.Tags {
				if tags[i] == draft.Tags[j].Value {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Tag %s not found in draft tags", tags[i])
			}
		}

		if len(draft.Tags) != 3 {
			t.Fatalf("Draft contains duplicate tags.  Expected %d, got %d", 3, len(draft.Tags))
		}

	})

	t.Run("New Draft", func(t *testing.T) {
		reset(t)
		draft, err := user.NewDocument("test draft title", "<h1>Test Draft Content</h1>",
			[]string{"tag1", "tag2", "tag3", "tag1", "tag1", "tag2"})
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

			err = d.Save(newTitle, newContent, newTags, 3)
			if !app.IsFail(err) {
				t.Fatalf("No failure on incorrect version: %s", err)
			}

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
				select count(*) from documents where id = {{arg "id"}}
			`).QueryRow(data.Arg("id", doc.ID)), 1)
			assertRow(t, data.NewQuery(`
				select count(*) from document_tags where document_id = {{arg "id"}}
			`).QueryRow(data.Arg("id", doc.ID)), 3)

			assertRow(t, data.NewQuery(`
				select count(*) from document_drafts where id = {{arg "id"}}
			`).QueryRow(data.Arg("id", draft.ID)), 0)

			assertRow(t, data.NewQuery(`
				select count(*) from document_draft_tags where draft_id = {{arg "id"}}
			`).QueryRow(data.Arg("id", draft.ID)), 0)

		})

	})

	t.Run("Existing Document Draft", func(t *testing.T) {
	})

}
