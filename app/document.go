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
	DraftID data.ID   `json:"draft_id"`
	Updated time.Time `json:"updated,omitempty"`
	Created time.Time `json:"created,omitempty"`
	creator data.ID
	updater data.ID
	groups  []data.ID

	DocumentContent

	accessor *User
}

// DocumentContent is the contents of a document who's structure is shared between drafts, history records, and
// published documents
type DocumentContent struct {
	ID      data.ID `json:"id"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Tags    []Tag   `json:"tags"`
	// Lanuage //TODO:
}

const (
	tagTypeUser = "user"
	tagTypeAuto = "auto"
)

// Tag is a string value that
type Tag struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Stem  string `json:"stem"`
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
	DraftID data.ID   `json:"draft_id"`
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
	deleteDraft,
	insert *data.Query
}{
	insert: data.NewQuery(`
		insert into documents (
			id,
			title,
			content,
			version,
			draft_id,
			updated,
			created,
			creator,
			updater
		) values (
			{{arg "id"}},
			{{arg "title"}},
			{{arg "content"}},
			{{arg "version"}},
			{{arg "draft_id"}},
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
			{{arg "tag"}},
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
			{{arg "tag"}},
			{{arg "type"}}
		)
	`),
	insertHistory: data.NewQuery(`
		insert into document_history (
			document_id,
			version,
			draft_id,
			title,
			content,
			created,
			creator,
		) values (
			{{arg "document_id"}},
			{{arg "version"}},
			{{arg "draft_id"}},
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
			draft_id = {{arg "draft_id"}},
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
				draft_id,
				updated,
				created,
				creator,
				updater
		from documents d
			inner join document_groups g on d.id = g.document_id
			inner join document_tags t on d.id = t.document_id
		where d.id = {{arg "id"}}
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
	deleteDraft: data.NewQuery(`
		delete from document_drafts
		where id = {{arg "id"}}
	`),
}

var (
	errDocumentConflict = Conflict("You are not editing the most current version of this document. " +
		"Please refresh and try again")
	errDocumentUpdateAccess = Unauthorized("You do not have access to update this document")
	errDocumentNotFound     = NotFound("Document not found")
)

var sanitizePolicy = bluemonday.UGCPolicy()

// Document retrieves a document
func DocumentGet(id data.ID, who *User) (*Document, error) {
	if id.IsNil() {
		return nil, errDocumentNotFound
	}

	d, err := documentGet(id)
	if err != nil {
		return nil, err
	}

	err = d.tryAccess(who)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func documentGet(id data.ID) (*Document, error) {
	d := &Document{}

	err := d.scan(sqlDocument.get.QueryRow(data.Arg("id", id)))
	if err != nil {
		return nil, err
	}
	// get tags
	// get groups
	//FIXME:
}

// tryAccess tries to access the document with the given user
func (d *Document) tryAccess(who *User) error {
	if len(d.groups) == 0 {
		if who == nil && !SettingMust("AllowPublicDocuments").Bool() {
			return errDocumentNotFound
		}
		d.accessor = who
		return nil
	}
	if who == nil {
		return errDocumentNotFound
	}

	return nil
}

// NewDocument starts a new document and returns the draft of it
func (u *User) NewDocument() (*DocumentDraft, error) {
	d := &DocumentDraft{
		ID:      data.NewID(),
		Version: 0,
		Updated: time.Now(),
		Created: time.Now(),
		creator: u.ID,
		updater: u.ID,
		editor:  u,
	}

	err := data.BeginTx(func(tx *sql.Tx) error {
		return d.insert(tx)
	})

	if err != nil {
		return nil, err
	}

	return d, nil
}

// NewDraft creates a new Draft for the given document
func (d *Document) NewDraft() (*DocumentDraft, error) {
	if d.accessor == nil {
		return nil, errDocumentUpdateAccess
	}
	draft := &DocumentDraft{
		ID:              data.NewID(),
		Version:         0,
		Updated:         time.Now(),
		Created:         time.Now(),
		creator:         d.accessor.ID,
		updater:         d.accessor.ID,
		editor:          d.accessor,
		DocumentContent: d.DocumentContent,
	}

	err := data.BeginTx(func(tx *sql.Tx) error {
		return draft.insert(tx)
	})

	if err != nil {
		return nil, err
	}

	return draft, nil
}

func (d *DocumentContent) validate() error {
	if strings.TrimSpace(d.Title) == "" {
		return NewFailure("A title is required on documents")
	}

	for i := range d.Tags {
		err := data.FieldValidate("document.tag", d.Tags[i].Value)
		if err != nil {
			return NewFailureFromErr(err)
		}
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
	return nil
}

// addTag adds a tag to the given document, including the tag's stem.  It won't add the tag if one already exists
func (d *DocumentContent) addTag(tagType, tagValue string) {
	for i := range d.Tags {
		if d.Tags[i].Value == tagValue && d.Tags[i].Type != tagTypeAuto {
			return
		}
	}

	tag := Tag{
		Value: tagValue,
		Type:  tagType,
	}

	tag.stem()

	d.Tags = append(d.Tags, tag)
}

func (t *Tag) stem() {
	//TODO: stemming
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

	for i := range d.Tags {
		_, err = sqlDocument.insertDraftTag.Tx(tx).Exec(
			data.Arg("draft_id", d.ID),
			data.Arg("tag", d.Tags[i].Value),
			data.Arg("type", d.Tags[i].Type),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Save saves the current document draft
func (d *DocumentDraft) Save(title, content string, tags []string, version int) error {
	return data.BeginTx(func(tx *sql.Tx) error {
		return d.update(tx, title, content, tags, version)
	})
}

func (d *DocumentDraft) update(tx *sql.Tx, title, content string, tags []string, version int) error {
	if d.editor == nil {
		return errDocumentUpdateAccess
	}
	d.Title = title
	d.Content = content
	d.Version = version
	d.Tags = nil

	for i := range tags {
		d.addTag(tagTypeUser, tags[i])
	}

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

	for i := range d.Tags {
		_, err = sqlDocument.insertDraftTag.Tx(tx).Exec(
			data.Arg("draft_id", d.ID),
			data.Arg("tag", d.Tags[i].Value),
			data.Arg("type", d.Tags[i].Type),
		)
		if err != nil {
			return err
		}

	}
	d.Version++

	return nil
}

func (d *Document) scan(record Scanner) error {
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

// make history turns the current document version into a history record
func (d *Document) makeHistory() *DocumentHistory {
	return &DocumentHistory{
		Version:         d.Version,
		DraftID:         d.DraftID,
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
func (d *DocumentDraft) Publish() (*Document, error) {
	if d.editor == nil {
		return nil, errDocumentUpdateAccess
	}

	var new *Document
	err := data.BeginTx(func(tx *sql.Tx) error {
		if d.DocumentContent.ID.IsNil() {
			// new document
			new = d.makeDocument(nil)
			err := new.insert(tx)

			if err != nil {
				return err
			}
		} else {
			old := &Document{}
			err := old.scan(sqlDocument.get.QueryRow(data.Arg("id", d.DocumentContent.ID)))
			if err != nil {
				return err
			}
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

		err := d.delete(tx)
		if err != nil {
			return err
		}

		err = new.link()
		if err != nil {
			return err
		}

		return new.index()
	})
	if err != nil {
		return nil, err
	}

	return new, nil
}

// makeDocument creates a document record from the current document draft
func (d *DocumentDraft) makeDocument(currentDocument *Document) *Document {
	if currentDocument == nil {

		return &Document{
			Version: 0,
			DraftID: d.ID,
			Updated: time.Now(),
			Created: time.Now(),
			creator: d.editor.ID,
			updater: d.editor.ID,
			DocumentContent: DocumentContent{
				ID:      data.NewID(),
				Title:   d.DocumentContent.Title,
				Content: d.DocumentContent.Content,
				Tags:    d.DocumentContent.Tags,
			},
		}
	}

	newDoc := *currentDocument
	newDoc.DraftID = d.ID
	newDoc.Updated = time.Now()
	newDoc.updater = d.editor.ID
	newDoc.DocumentContent = d.DocumentContent

	return &newDoc
}

func (d *DocumentDraft) delete(tx *sql.Tx) error {
	if tx == nil {
		panic("A transaction is required when deleting a draft")
	}
	_, err := sqlDocument.deleteDraftTags.Tx(tx).Exec(data.Arg("draft_id", d.ID))
	if err != nil {
		return err
	}
	_, err = sqlDocument.deleteDraft.Tx(tx).Exec(data.Arg("id", d.ID))

	return err
}

func (h *DocumentHistory) insert(tx *sql.Tx) error {
	_, err := sqlDocument.insertHistory.Tx(tx).Exec(
		data.Arg("document_id", h.ID),
		data.Arg("version", h.Version),
		data.Arg("draft_id", h.DraftID),
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
		data.Arg("draft_id", d.DraftID),
		data.Arg("updated", d.Updated),
		data.Arg("created", d.Created),
		data.Arg("creator", d.creator),
		data.Arg("updater", d.updater),
	)
	if err != nil {
		return err
	}

	for i := range d.Tags {
		_, err = sqlDocument.insertTag.Tx(tx).Exec(
			data.Arg("document_id", d.ID),
			data.Arg("tag", d.Tags[i].Value),
			data.Arg("type", d.Tags[i].Type),
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
		data.Arg("draft_id", d.DraftID),
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

	for i := range d.Tags {
		_, err = sqlDocument.insertTag.Tx(tx).Exec(
			data.Arg("document_id", d.ID),
			data.Arg("tag", d.Tags[i].Value),
			data.Arg("type", d.Tags[i].Type),
		)
		if err != nil {
			return err
		}

	}

	return nil
}
