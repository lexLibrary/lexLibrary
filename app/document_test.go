// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"golang.org/x/text/language"
)

func TestDocument(t *testing.T) {
	var user *app.User
	reset := func(t *testing.T) {
		t.Helper()

		user = resetAdmin(t, "admin", "adminpassword").User()
	}

	t.Run("New Document", func(t *testing.T) {
		reset(t)

		draft, err := user.NewDocument(language.English)
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
		draft, err := user.NewDocument(language.English)
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

		t.Run("Publish", func(t *testing.T) {
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
				select document_id, language, version, title, content 
				from document_contents 
				where document_id = {{arg "id"}}
				and language = {{arg "language"}}
			`).QueryRow(data.Arg("id", doc.ID), data.Arg("language", draft.Language.String())),
				doc.ID, doc.Language.String(), doc.Version, draft.Title, draft.Content)
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
				where document_id = {{arg "document_id"}}
			`).QueryRow(data.Arg("document_id", doc.ID)), 0)
		})

		t.Run("Get", func(t *testing.T) {

		})

	})

	t.Run("Existing Document", func(t *testing.T) {
		reset(t)
		draft, err := user.NewDocument(language.English)
		if err != nil {
			t.Fatal(err)
		}

		draft.Save("Title", "<h2>Content</h2>", []string{"tag1", "tag2", "tag3"}, draft.Version)
		doc, err := draft.Publish()
		if err != nil {
			t.Fatal(err)
		}

		t.Run("New Draft", func(t *testing.T) {
			first, err := doc.NewDraft(doc.Language)
			if err != nil {
				t.Fatalf("Error creating new draft of existing document: %s", err)
			}

			assertRow(t, data.NewQuery(`
				select count(*) from document_drafts 
				where document_id = {{arg "document_id"}}
				and language = {{arg "language"}}
				and id = {{arg "id"}}
			`).QueryRow(
				data.Arg("document_id", doc.ID),
				data.Arg("language", doc.Language.String()),
				data.Arg("id", first.ID),
			), 1)

			assertRow(t, data.NewQuery(`
				select count(*) from document_draft_tags
				where language = {{arg "language"}}
				and draft_id = {{arg "id"}}
			`).QueryRow(
				data.Arg("language", doc.Language.String()),
				data.Arg("id", first.ID),
			), 3)

			second, err := doc.NewDraft(doc.Language)
			if err != nil {
				t.Fatalf("Error creating a second draft of existing document: %s", err)
			}
			if second.ID == first.ID {
				t.Fatal("Second draft has matching draft id to first draft")
			}

			assertRow(t, data.NewQuery(`
				select count(*) from document_drafts 
				where document_id = {{arg "document_id"}}
				and language = {{arg "language"}}
				and id = {{arg "id"}}
			`).QueryRow(
				data.Arg("document_id", doc.ID),
				data.Arg("language", doc.Language.String()),
				data.Arg("id", second.ID),
			), 1)

			assertRow(t, data.NewQuery(`
				select count(*) from document_draft_tags
				where language = {{arg "language"}}
				and draft_id = {{arg "id"}}
			`).QueryRow(
				data.Arg("language", doc.Language.String()),
				data.Arg("id", second.ID),
			), 3)

			assertRow(t, data.NewQuery(`
				select count(*) from document_drafts 
				where document_id = {{arg "document_id"}}
				and language = {{arg "language"}}
			`).QueryRow(
				data.Arg("document_id", doc.ID),
				data.Arg("language", doc.Language.String()),
			), 2)

			third, err := doc.NewDraft(language.Polish)
			if err != nil {
				t.Fatalf("Error creating a third draft of existing document in a new language: %s", err)
			}

			if third.Language.String() == doc.Language.String() {
				t.Fatalf("New language draft doesn't have a different language from the original")
			}

			assertRow(t, data.NewQuery(`
				select count(*) from document_drafts 
				where document_id = {{arg "document_id"}}
				and language = {{arg "language"}}
				and id = {{arg "id"}}
			`).QueryRow(
				data.Arg("document_id", doc.ID),
				data.Arg("language", third.Language.String()),
				data.Arg("id", third.ID),
			), 1)

			assertRow(t, data.NewQuery(`
				select count(*) from document_draft_tags
				where language = {{arg "language"}}
				and draft_id = {{arg "id"}}
			`).QueryRow(
				data.Arg("language", third.Language.String()),
				data.Arg("id", third.ID),
			), 3)

			assertRow(t, data.NewQuery(`
				select count(*) from document_drafts 
				where document_id = {{arg "document_id"}}
				and language = {{arg "language"}}
			`).QueryRow(
				data.Arg("document_id", doc.ID),
				data.Arg("language", third.Language.String()),
			), 1)

			firstTitle := "First Title"
			secondTitle := "Second Title"
			thirdTitle := "Third Title"
			firstContent := "<h3>First<h3>"
			secondContent := "<h3>2nd<h3>"
			thirdContent := "<h3>Third<h3>"
			firstTags := []string{"1", "one", "first"}
			secondTags := []string{"2", "two", "second", "too"}
			thirdTags := []string{"3", "three"}

			err = first.Save(firstTitle, firstContent, firstTags, first.Version)
			if err != nil {
				t.Fatal(err)
			}

			err = second.Save(secondTitle, secondContent, secondTags, second.Version)
			if err != nil {
				t.Fatal(err)
			}

			err = third.Save(thirdTitle, thirdContent, thirdTags, third.Version)
			if err != nil {
				t.Fatal(err)
			}

			draftQuery := data.NewQuery(`
				select title, content, d.language, count(*)
				from document_drafts d
					inner join document_draft_tags t on d.id = t.draft_id
				where d.id = {{arg "id"}}
				group by title, content, d.language
			`)

			assertRow(t, draftQuery.QueryRow(data.Arg("id", first.ID)),
				firstTitle, firstContent, first.Language.String(), 3)
			assertRow(t, draftQuery.QueryRow(data.Arg("id", second.ID)),
				secondTitle, secondContent, second.Language.String(), 4)
			assertRow(t, draftQuery.QueryRow(data.Arg("id", third.ID)),
				thirdTitle, thirdContent, third.Language.String(), 2)

			publishQuery := data.NewQuery(`
				select tbl1.documents, tbl2.drafts
				from (
					select count(*) as documents
					from document_contents
					where document_id = {{arg "id"}}
				)tbl1,
				(
					select count(*) as drafts
					from document_drafts
					where document_id = {{arg "document_id"}}
				)tbl2
			`)

			_, err = first.Publish()
			if err != nil {
				t.Fatal(err)
			}

			assertRow(t, publishQuery.QueryRow(data.Arg("id", doc.ID), data.Arg("document_id", doc.ID)),
				1, 2)

			_, err = second.Publish()
			if err != nil {
				t.Fatal(err)
			}

			assertRow(t, publishQuery.QueryRow(data.Arg("id", doc.ID), data.Arg("document_id", doc.ID)),
				1, 1)
			_, err = third.Publish()
			if err != nil {
				t.Fatal(err)
			}

			assertRow(t, publishQuery.QueryRow(data.Arg("id", doc.ID), data.Arg("document_id", doc.ID)),
				2, 0)

		})

		t.Run("Groups", func(t *testing.T) {
			group1, err := user.NewGroup("test document Group")
			if err != nil {
				t.Fatal(err)
			}
			group2, err := user.NewGroup("test document Group 2")
			if err != nil {
				t.Fatal(err)
			}

			t.Run("Add", func(t *testing.T) {
				err = doc.AddGroup(data.ID{})
				if !app.IsFail(err) {
					t.Fatalf("Adding null group ID to document did not fail: %s", err)
				}

				err = doc.AddGroup(data.NewID())
				if !app.IsFail(err) {
					t.Fatalf("Adding invalid group to document did not fail: %s", err)
				}

				err = doc.AddGroup(group1.ID)
				if err != nil {
					t.Fatal(err)
				}
				assertQuery := data.NewQuery(`
					select count(*) 
					from document_groups 
					where document_id = {{arg "document_id"}}
				`)

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 1)

				err = doc.AddGroup(group2.ID)
				if err != nil {
					t.Fatal(err)
				}

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 2)

				err = doc.AddGroup(group1.ID)
				if err != nil {
					t.Fatalf("Error adding a group that already exists: %s", err)
				}
			})

			t.Run("Remove", func(t *testing.T) {
				err = doc.RemoveGroup(data.ID{})
				if !app.IsFail(err) {
					t.Fatalf("Removing null group ID from document did not fail: %s", err)
				}

				err = doc.RemoveGroup(data.NewID())
				if err != nil {
					t.Fatalf("Removing invalid group from document failed: %s", err)
				}

				err = doc.RemoveGroup(group1.ID)
				if err != nil {
					t.Fatal(err)
				}
				assertQuery := data.NewQuery(`
					select count(*) 
					from document_groups 
					where document_id = {{arg "document_id"}}
				`)

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 1)

				err = doc.RemoveGroup(group2.ID)
				if err != nil {
					t.Fatal(err)
				}

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 0)

				err = doc.RemoveGroup(group1.ID)
				if err != nil {
					t.Fatalf("Error removing a group that is already gone: %s", err)
				}
			})
		})

		t.Run("Get", func(t *testing.T) {

		})
	})

}
