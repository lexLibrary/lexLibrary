// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"net/http"
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
		ok(t, err)

		assert(t, draft != nil, "Draft is nil")
		equals(t, 0, draft.Version)
		assert(t, !draft.ID.IsNil(), "Draft ID is nil")
		assert(t, draft.DocumentContent.ID.IsNil(), "Document ID is not nil on new document")

		assertRow(t, data.NewQuery(`
			select count(*) from document_drafts where id = {{arg "id"}}
		`).QueryRow(data.Arg("id", draft.ID)), 1)

	})

	t.Run("New Draft", func(t *testing.T) {
		reset(t)
		draft, err := user.NewDocument(language.English)
		ok(t, err)

		t.Run("Save", func(t *testing.T) {
			newTitle := "new Title"
			newContent := "<h2>New Content</h2>"
			newTags := []string{"newtag1", "newtag2", "newtag3", "newtag1", "newtag1", "newtag2"}

			d := *draft

			assertFail(t, d.Save("", newContent, newTags, draft.Version), http.StatusBadRequest,
				"Empty title did not fail")

			d = *draft
			assertFail(t, d.Save(newTitle, newContent, newTags, 3), http.StatusConflict,
				"No failure on incorrect version")

			d = *draft
			assertFail(t, d.Save(newTitle, newContent, []string{fmt.Sprintf("Long %70s", "tag value")}, 3),
				http.StatusBadRequest, "No failure on long tag")

			ok(t, draft.Save(newTitle, newContent, newTags, draft.Version))
			equals(t, 1, draft.Version)
			equals(t, newTitle, draft.Title)
			equals(t, newContent, draft.Content)

			for i := range newTags {
				found := false
				for j := range draft.Tags {
					if newTags[i] == draft.Tags[j].Value {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("Tag %s not found in draft tags", newTags[i])
				}
			}

			equals(t, 3, len(draft.Tags))

			t.Run("Publish", func(t *testing.T) {
				// test publishing an empty draft
				empty, err := user.NewDocument(language.English)
				ok(t, err)
				_, err = empty.Publish()
				assertFail(t, err, http.StatusBadRequest, "Didn't fail on publishing an empty draft")

				doc, err := draft.Publish()
				ok(t, err)

				assert(t, doc != nil, "Published Document is nil")

				equals(t, draft.Title, doc.Title)
				equals(t, draft.Content, doc.Content)

				assert(t, !doc.ID.IsNil(), "Doc has nil ID")

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
			`).QueryRow(data.Arg("id", doc.ID)), 3)

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

		})
	})

	t.Run("Existing Document", func(t *testing.T) {
		reset(t)
		draft, err := user.NewDocument(language.English)
		ok(t, err)

		ok(t, draft.Save("Title", "<h2>Content</h2>", []string{"tag1", "tag2", "tag3"}, draft.Version))
		doc, err := draft.Publish()
		ok(t, err)

		t.Run("New Draft", func(t *testing.T) {
			first, err := doc.NewDraft(doc.Language)
			ok(t, err)
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
			ok(t, err)
			assert(t, first.ID != second.ID, "Second draft has matching draft id to first draft")

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
			ok(t, err)

			assert(t, third.Language.String() != doc.Language.String(),
				"New language draft doesn't have a different language from the original")

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
			ok(t, err)

			err = second.Save(secondTitle, secondContent, secondTags, second.Version)
			ok(t, err)

			err = third.Save(thirdTitle, thirdContent, thirdTags, third.Version)
			ok(t, err)

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
			ok(t, err)

			assertRow(t, publishQuery.QueryRow(data.Arg("id", doc.ID), data.Arg("document_id", doc.ID)),
				1, 2)

			_, err = second.Publish()
			ok(t, err)

			assertRow(t, publishQuery.QueryRow(data.Arg("id", doc.ID), data.Arg("document_id", doc.ID)),
				1, 1)
			_, err = third.Publish()
			ok(t, err)

			assertRow(t, publishQuery.QueryRow(data.Arg("id", doc.ID), data.Arg("document_id", doc.ID)),
				2, 0)

		})

		t.Run("Groups", func(t *testing.T) {
			group1, err := user.NewGroup("test document Group")
			ok(t, err)
			group2, err := user.NewGroup("test document Group 2")
			ok(t, err)

			t.Run("Add", func(t *testing.T) {
				assertFail(t, doc.AddGroup(data.ID{}, false), http.StatusBadRequest,
					"Adding null group ID to document did not fail")

				assertFail(t, doc.AddGroup(data.NewID(), false), http.StatusNotFound,
					"Adding invalid group to document did not fail")

				ok(t, doc.AddGroup(group1.ID, false))
				assertQuery := data.NewQuery(`
					select count(*) 
					from document_groups 
					where document_id = {{arg "document_id"}}
				`)

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 1)

				err = doc.AddGroup(group2.ID, false)
				if err != nil {
					t.Fatal(err)
				}

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 2)

				ok(t, doc.AddGroup(group1.ID, true))
				assertRow(t, data.NewQuery(`
					select can_publish
					from document_groups
					where document_id = {{arg "document_id"}}
					and group_id = {{arg "group_id"}}
				`).QueryRow(data.Arg("document_id", doc.ID), data.Arg("group_id", group1.ID)), true)
			})

			t.Run("Remove", func(t *testing.T) {
				assertFail(t, doc.RemoveGroup(data.ID{}), http.StatusBadRequest,
					"Removing null group ID from document did not fail")

				ok(t, doc.RemoveGroup(data.NewID()))

				ok(t, doc.RemoveGroup(group1.ID))
				assertQuery := data.NewQuery(`
					select count(*) 
					from document_groups 
					where document_id = {{arg "document_id"}}
				`)

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 1)

				ok(t, doc.RemoveGroup(group2.ID))

				assertRow(t, assertQuery.QueryRow(data.Arg("document_id", doc.ID)), 0)

				ok(t, doc.RemoveGroup(group1.ID))
			})

			t.Run("Permissions", func(t *testing.T) {
				admin, err := user.Admin()
				ok(t, err)
				ok(t, admin.SetSetting("AllowPublicDocuments", true))
				ok(t, admin.SetSetting("AllowPublicSignups", true))

				other, err := app.UserNew("otheruser", "otheruserpassword")
				ok(t, err)

				doc, err := app.DocumentGet(doc.ID, doc.Language, nil)
				ok(t, err)

				assertFail(t, doc.AddGroup(group1.ID, false), http.StatusUnauthorized,
					"Adding a group to a document accessed publically")
				assertFail(t, doc.RemoveGroup(group1.ID), http.StatusUnauthorized,
					"Removing a group to a document accessed publically")

				doc, err = app.DocumentGet(doc.ID, doc.Language, other)
				ok(t, err)

				assertFail(t, doc.AddGroup(group1.ID, false), http.StatusUnauthorized,
					"Adding a group to a document accessed by a non-owner")

				assertFail(t, doc.RemoveGroup(group1.ID), http.StatusUnauthorized,
					"Removing a group to a document accessed by a non-owner")
			})
		})
	})

	t.Run("Get", func(t *testing.T) {
		reset(t)

		draft, err := user.NewDocument(language.English)
		ok(t, err)
		ok(t, draft.Save("Title", "<h1>Content</h1>", []string{"tag1", "tag2", "tag3", "tag4"}, draft.Version))

		doc, err := draft.Publish()
		ok(t, err)

		admin, err := user.Admin()
		ok(t, err)
		ok(t, admin.SetSetting("AllowPublicDocuments", false))

		_, err = app.DocumentGet(data.ID{}, language.English, user)
		assertFail(t, err, http.StatusNotFound, "Getting document with nil ID did not fail")

		_, err = app.DocumentGet(data.NewID(), language.English, user)
		assertFail(t, err, http.StatusNotFound, "Getting document with an invalid ID did not fail")

		_, err = app.DocumentGet(doc.ID, language.Polish, user)
		assertFail(t, err, http.StatusNotFound, "Getting document with incorrect language")

		_, err = app.DocumentGet(doc.ID, language.English, nil)
		assertFail(t, err, http.StatusNotFound, "Getting private document with no user")

		ok(t, admin.SetSetting("AllowPublicDocuments", true))

		other, err := app.DocumentGet(doc.ID, language.English, nil)
		ok(t, err)

		equals(t, doc.Title, other.Title)
		equals(t, doc.Content, other.Content)
		equals(t, doc.Tags, other.Tags)

		// doc with no tags
		newDraft, err := user.NewDocument(language.Ukrainian)
		ok(t, err)
		ok(t, newDraft.Save("Title", "<h1>Content</h1>", nil, newDraft.Version))

		newDoc, err := newDraft.Publish()
		ok(t, err)

		other, err = app.DocumentGet(newDoc.ID, language.Ukrainian, nil)
		ok(t, err)
		equals(t, newDoc.Title, other.Title)
		equals(t, newDoc.Content, other.Content)
		equals(t, newDoc.Tags, other.Tags)

		t.Run("Group Access", func(t *testing.T) {
			ok(t, admin.SetSetting("AllowPublicSignups", true))
			// otherUser, err := app.UserNew("other", "otherPassword")
			// ok(t, err)

			blue, err := user.NewGroup("blue")
			ok(t, err)
			// red, err := user.NewGroup("red")
			// ok(t, err)

			ok(t, doc.AddGroup(blue.ID, false))
		})

	})

	t.Run("Get Draft", func(t *testing.T) {

	})
}
