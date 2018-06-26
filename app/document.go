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
	tags    []Tag
}

const (
	tagTypeUser = "user"
	tagTypeAuto = "auto"
)

// Tag is a string value that
type Tag struct {
	Value string `json: "value"`
	Type  string `json: "type"`
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

	editor *User // current user editing the draft
}

// DocumentHistory is a previously published version of a document
type DocumentHistory struct {
	Version int       `json:"version"`
	Created time.Time `json:"created,omitempty"`
	creator data.ID

	DocumentContent
}

var sqlDocument = struct {
	insertGroup,
	insertTag,
	insertDraft,
	insertDraftTag,
	insertHistory,
	updateDraft,
	get,
	getDraftTags,
	getTags,
	update,
	deleteTags,
	deleteDraftTags,
	insert *data.Query
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
			tag,
			type
		) values (
			{{arg "document_id"}},
			{{arg "tag"}}
			{{arg "type"}}
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
			tag,
			type
		) values (
			{{arg "draft_id"}},
			{{arg "tag"}}
			{{arg "type"}}
		)
	`),
	insertHistory: data.NewQuery(`
		insert into document_history (
			document_id,
			version,
			title,
			content,
			created,
			creator,
		) values (
			{{arg "document_id"}},
			{{arg "version"}},
			{{arg "title"}},
			{{arg "content"}},
			{{arg "created"}},
			{{arg "creator"}},
		)
	`),
	updateDraft: data.NewQuery(`
		update document_drafts 
		set updated = {{NOW}}, 
			version = version + 1,
			updater = {{arg "updater"}},
			title = {{arg "title"}},
			content = {{arg "content"}}
		where id = {{arg "id"}} 
		and version = {{arg "version"}}
	`),
	update: data.NewQuery(`
		update documents 
		set updated = {{NOW}}, 
			version = version + 1,
			updater = {{arg "updater"}},
			title = {{arg "title"}},
			content = {{arg "content"}}
		where id = {{arg "id"}} 
		and version = {{arg "version"}}
	`),
	get: data.NewQuery(`
		select 	id,
			title,
			content,
			version,
			updated,
			created,
			creator,
			updater
		from documents
		where id = {{arg "id"}}
	`),
	getTags: data.NewQuery(`
		select 	document_id,
			tag,
			type
		from document_tags
		where	document_id = {{arg "id"}}	
	`),
	deleteTags: data.NewQuery(`
		delete from document_tags
		where document_id = {{arg "document_id"}}
	`),
	deleteDraftTags: data.NewQuery(`
		delete from document_draft_tags
		where draft_id = {{arg "draft_id"}}
	`),
}

var (
	errDocumentConflict = Conflict("You are not editing the most current version of this document. " +
		"Please refresh and try again")
	errDocumentUpdateAccess = Unauthorized("You do not have access to update this document")
	errDocumentNotFound     = NotFound("Document not found")
)

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
		},
		editor: u,
	}

	d.mergeTags(tagTypeUser, tags...)

	err := d.validate()
	if err != nil {
		return nil, err
	}

	d.sanitize()

	err = d.autoTag()
	if err != nil {
		return nil, err
	}

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

// autoTag builds tags automatically based on the document's content. User specified tags
// have a greater weight than auto generated tags
func (d *DocumentContent) autoTag() error {
	//TODO: don't overwrite any user tags

	return nil
}

// mergeTags merges the passed in tags with the current document
func (d *DocumentContent) mergeTags(tagType string, tagList ...string) {

	for i := range tagList {
		found := false
		for j := range d.tags {
			if d.tags[j].Value == tagList[i] {
				found = true
				break
			}
		}
		if !found {
			d.tags = append(d.tags, Tag{
				Value: tagList[i],
				Type:  tagType,
			})
		}
	}
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
			data.Arg("tag", d.tags[i].Value),
			data.Arg("type", d.tags[i].Type),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DocumentDraft) canEdit() bool {
	// TODO: Invite others to edit your draft
	return d.editor != nil && d.creator == d.editor.ID
}

// Save saves the current document draft
func (d *DocumentDraft) Save(title, content string, tags []string, version int) error {
	return data.BeginTx(func(tx *sql.Tx) error {
		return d.update(tx, title, content, tags, version)
	})
}

func (d *DocumentDraft) update(tx *sql.Tx, title, content string, tags []string, version int) error {
	if !d.canEdit() {
		return errDocumentUpdateAccess
	}
	d.Title = title
	d.Content = content
	d.mergeTags(tagTypeUser, tags...)

	err := d.validate()
	if err != nil {
		return err
	}

	d.sanitize()

	err = d.autoTag()
	if err != nil {
		return err
	}

	r, err := sqlDocument.updateDraft.Tx(tx).Exec(
		data.Arg("id", d.ID),
		data.Arg("version", d.Version),
		data.Arg("updater", d.editor.ID),
		data.Arg("title", d.Title),
		data.Arg("content", d.Content),
	)

	if err != nil {
		return err
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errDocumentConflict
	}

	_, err = sqlDocument.deleteDraftTags.Tx(tx).Exec(data.Arg("draft_id", d.ID))
	if err != nil {
		return err
	}

	for i := range d.tags {
		_, err = sqlDocument.insertDraftTag.Tx(tx).Exec(
			data.Arg("draft_id", d.ID),
			data.Arg("tag", d.tags[i].Value),
			data.Arg("type", d.tags[i].Type),
		)
		if err != nil {
			return err
		}

	}

	return nil
}

func (d *Document) scan(record scanner) error {
	err := record.Scan(
		&d.ID,
		&d.Title,
		&d.Content,
		&d.Version,
		&d.Updated,
		&d.Created,
		&d.creator,
		&d.updater,
	)
	if err == sql.ErrNoRows {
		return errDocumentNotFound
	}
	return err
}

// Tags returns the tags for the given document
// func (d *DocumentContent) Tags() ([]Tag, error) {

// }

// make history turns the current document version into a history record
func (d *Document) makeHistory() *DocumentHistory {
	return &DocumentHistory{
		Version:         d.Version,
		Created:         d.Updated, // history created is current updated
		creator:         d.updater, // history creator is current updater
		DocumentContent: d.DocumentContent,
	}
}

// link builds weighted links to other published documents based on their tags
func (d *Document) link() error {
	//TODO:
	return nil
}

// index adds the document to the search index
func (d *Document) index() error {
	//TODO:
	return nil
}

// Publish publishes a draft turing a draft into a document
func (d *DocumentDraft) Publish() error {
	if !d.canEdit() {
		return errDocumentUpdateAccess
	}

	return data.BeginTx(func(tx *sql.Tx) error {
		var new *Document
		old := &Document{}
		err := old.scan(sqlDocument.get.QueryRow(data.Arg("id", d.DocumentContent.ID)))
		if err == errDocumentNotFound {
			new = d.makeDocument(nil)
			err = new.insert(tx)

			if err != nil {
				return err
			}
		} else {
			err = old.makeHistory().insert(tx)
			if err != nil {
				return err
			}
			new = d.makeDocument(old)
			err = new.update(tx)

			if err != nil {
				return err
			}

		}

		err = new.link()
		if err != nil {
			return err
		}

		return new.index()
	})
}

// makeDocument creates a document record from the current document draft
func (d *DocumentDraft) makeDocument(currentDocument *Document) *Document {
	if currentDocument == nil {
		return &Document{
			Version:         0,
			Updated:         time.Now(),
			Created:         time.Now(),
			creator:         d.editor.ID,
			updater:         d.editor.ID,
			DocumentContent: d.DocumentContent,
		}
	}

	newDoc := *currentDocument
	newDoc.Updated = time.Now()
	newDoc.updater = d.editor.ID
	newDoc.DocumentContent = d.DocumentContent

	return &newDoc
}

func (h *DocumentHistory) insert(tx *sql.Tx) error {
	_, err := sqlDocument.insertHistory.Tx(tx).Exec(
		data.Arg("document_id", h.ID),
		data.Arg("version", h.Version),
		data.Arg("title", h.Title),
		data.Arg("content", h.Content),
		data.Arg("created", h.Created),
		data.Arg("creator", h.creator),
	)

	return err
}

func (d *Document) insert(tx *sql.Tx) error {
	if tx == nil {
		panic("A transaction is required when inserting a document")
	}

	_, err := sqlDocument.insert.Tx(tx).Exec(
		data.Arg("id", d.ID),
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
		_, err = sqlDocument.insertTag.Tx(tx).Exec(
			data.Arg("draft_id", d.ID),
			data.Arg("tag", d.tags[i].Value),
			data.Arg("type", d.tags[i].Type),
		)
		if err != nil {
			return err
		}
	}
	return nil

}

func (d *Document) update(tx *sql.Tx) error {
	if tx == nil {
		panic("A transaction is required when updating a document")
	}

	r, err := sqlDocument.update.Tx(tx).Exec(
		data.Arg("id", d.ID),
		data.Arg("version", d.Version),
		data.Arg("updater", d.updater),
		data.Arg("title", d.Title),
		data.Arg("content", d.Content),
	)

	if err != nil {
		return err
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errDocumentConflict
	}

	_, err = sqlDocument.deleteTags.Tx(tx).Exec(data.Arg("document_id", d.ID))
	if err != nil {
		return err
	}

	for i := range d.tags {
		_, err = sqlDocument.insertTag.Tx(tx).Exec(
			data.Arg("document_id", d.ID),
			data.Arg("tag", d.tags[i].Value),
			data.Arg("type", d.tags[i].Type),
		)
		if err != nil {
			return err
		}

	}

	return nil
}
