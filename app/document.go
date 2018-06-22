// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/microcosm-cc/bluemonday"
)

// Document is an instance of a published document
type Document struct {
	Version int       `json:"version"`
	Updated time.Time `json:"updated,omitempty"`
	Created time.Time `json:"created,omitempty"`
	creator data.ID
	updater data.ID
	groups  []data.ID

	DocumentContent
}

// DocumentContent is the contents of a document who's structure is shared between drafts, history records, and
// published documents
type DocumentContent struct {
	ID      data.ID `json:"id"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	tags    []string
}

// DocumentDraft is a draft of a document, not visible to the public
type DocumentDraft struct {
	ID      data.ID   `json:"id"`
	Version int       `json:"version"`
	Updated time.Time `json:"updated,omitempty"`
	Created time.Time `json:"created,omitempty"`
	creator data.ID
	updater data.ID

	DocumentContent
}

// DocumentHistory is a previously published version of a document
type DocumentHistory struct {
	draftID data.ID   `json:"draft_id"`
	Created time.Time `json:"created,omitempty"`
	creator data.ID

	DocumentContent
}

var sqlDocument = struct {
	insert,
	insertGroup,
	insertTag,
	insertDraft,
	insertDraftTag,
	insertHistory *data.Query
}{
	insert: data.NewQuery(`
		insert into documents (
			id,
			title,
			content,
			version,
			updated,
			created,
			creator,
			updater
		) values (
			{{arg "id"}},
			{{arg "title"}},
			{{arg "content"}},
			{{arg "version"}},
			{{arg "updated"}},
			{{arg "created"}},
			{{arg "creator"}},
			{{arg "updater"}}
		)
	`),
	insertGroup: data.NewQuery(`
		insert into document_groups (
			document_id,
			group_id
		) values (
			{{arg "document_id"}},
			{{arg "group_id"}}
		)
	`),
	insertTag: data.NewQuery(`
		insert into document_tags (
			document_id,
			tag
		) values (
			{{arg "document_id"}},
			{{arg "tag"}}
		)
	`),
	insertDraft: data.NewQuery(`
		insert into document_drafts (
			id,
			document_id,
			title,
			content,
			version,
			updated,
			created,
			creator,
			updater
		) values (
			{{arg "id"}},
			{{arg "document_id"}},
			{{arg "title"}},
			{{arg "content"}},
			{{arg "version"}},
			{{arg "updated"}},
			{{arg "created"}},
			{{arg "creator"}},
			{{arg "updater"}}
		)
	`),
	insertDraftTag: data.NewQuery(`
		insert into document_draft_tags (
			draft_id,
			tag
		) values (
			{{arg "draft_id"}},
			{{arg "tag"}}
		)
	`),
	insertHistory: data.NewQuery(`
		insert into document_history (
			document_id,
			version,
			title,
			content,
			draft_id,
			created,
			creator,
		) values (
			{{arg "document_id"}},
			{{arg "version"}},
			{{arg "title"}},
			{{arg "content"}},
			{{arg "draft_id"}},
			{{arg "created"}},
			{{arg "creator"}},
		)
	`),
}

var sanitizePolicy = bluemonday.UGCPolicy()

// NewDocument starts a new document and returns the draft of it
func (u *User) NewDocument(title, content string, tags []string, groups []data.ID) (*DocumentDraft, error) {
	d := &DocumentDraft{
		ID:      data.NewID(),
		Version: 0,
		Updated: time.Now(),
		Created: time.Now(),
		creator: u.ID,
		updater: u.ID,
		DocumentContent: DocumentContent{
			ID:      data.NewID(),
			Title:   title,
			Content: content,
			tags:    tags,
		},
	}

	err := d.validate()
	if err != nil {
		return nil, err
	}

	d.sanitize()

	err = data.BeginTx(func(tx *sql.Tx) error {
		return d.insert(tx)
	})

	if err != nil {
		return nil, err
	}

	return d, nil

}

func (d *DocumentContent) validate() error {
	if strings.TrimSpace(d.Title) == "" {
		return NewFailure("A title is required on documents")
	}

	return nil
}

// sanitize removes any unneeded, unsupported, or unsafe content
func (d *DocumentContent) sanitize() {
	d.Content = sanitizePolicy.Sanitize(d.Content)
}

func (d *DocumentDraft) insert(tx *sql.Tx) error {
	if tx == nil {
		panic("A transaction is required when inserting a document draft")
	}

	_, err := sqlDocument.insertDraft.Tx(tx).Exec(
		data.Arg("id", d.ID),
		data.Arg("document_id", d.DocumentContent.ID),
		data.Arg("title", d.Title),
		data.Arg("content", d.Content),
		data.Arg("version", d.Version),
		data.Arg("updated", d.Updated),
		data.Arg("created", d.Created),
		data.Arg("creator", d.creator),
		data.Arg("updater", d.updater),
	)
	if err != nil {
		return err
	}

	for i := range d.tags {
		_, err = sqlDocument.insertDraftTag.Tx(tx).Exec(
			data.Arg("draft_id", d.ID),
			data.Arg("tag", d.tags[i]),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// func (d *DocumentDraft) Update(title, content string, version int) error {

// }

// Publish publishes a draft turing a draft into a document
// func (d *DocumentDraft) Publish() (*Document, error) {

// }
