package app

import "github.com/lexLibrary/lexLibrary/data"

var sqlDocument = struct {
	insertGroup,
	insertTag,
	insertDraft,
	insertDraftTag,
	insertHistory,
	updateDraft,
	get,
	getDraft,
	update,
	deleteGroup,
	deleteTags,
	deleteDraftTags,
	deleteDraft,
	insertContent,
	groupExists,
	updateGroup,
	canAccessDraft,
	canPublish,
	insert *data.Query
}{
	insert: data.NewQuery(`
		insert into documents (
			id,
			created,
			creator
		) values (
			{{arg "id"}},
			{{arg "created"}},
			{{arg "creator"}}
		)
	`),
	insertContent: data.NewQuery(`
		insert into document_contents (
			document_id,
			language,
			version,
			title,
			content,
			created,
			creator,
			updated,
			updater
		) values (
			{{arg "document_id"}},
			{{arg "language"}},
			{{arg "version"}},
			{{arg "title"}},
			{{arg "content"}},
			{{arg "created"}},
			{{arg "creator"}},
			{{arg "updated"}},
			{{arg "updater"}}
		)
	`),
	insertGroup: data.NewQuery(`
		insert into document_groups (
			document_id,
			group_id,
			can_publish
		) select
			{{arg "document_id"}},
			id,
			{{arg "can_publish"}}
		from	groups
		where id = {{arg "group_id"}}
	`),
	insertTag: data.NewQuery(`
		insert into document_tags (
			document_id,
			language,
			tag,
			stem,
			type
		) values (
			{{arg "document_id"}},
			{{arg "language"}},
			{{arg "tag"}},
			{{arg "stem"}},
			{{arg "type"}}
		)
	`),
	insertDraft: data.NewQuery(`
		insert into document_drafts (
			id,
			document_id,
			language,
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
			{{arg "language"}},
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
			language,
			tag,
			stem,
			type
		) values (
			{{arg "draft_id"}},
			{{arg "language"}},
			{{arg "tag"}},
			{{arg "stem"}},
			{{arg "type"}}
		)
	`),
	insertHistory: data.NewQuery(`
		insert into document_history (
			document_id,
			language,
			version,
			title,
			content,
			created,
			creator
		) values (
			{{arg "document_id"}},
			{{arg "language"}},
			{{arg "version"}},
			{{arg "title"}},
			{{arg "content"}},
			{{arg "created"}},
			{{arg "creator"}}
		)
	`),
	updateDraft: data.NewQuery(`
		update document_drafts 
		set 	updated = {{NOW}}, 
			version = version + 1,
			updater = {{arg "updater"}},
			title = {{arg "title"}},
			content = {{arg "content"}}
		where id = {{arg "id"}} 
		and version = {{arg "version"}}
		and language = {{arg "language"}}
	`),
	update: data.NewQuery(`
		update document_contents 
		set 	title = {{arg "title"}},
			content = {{arg "content"}},
			updated = {{NOW}}, 
			version = version + 1,
			updater = {{arg "updater"}}
		where document_id = {{arg "document_id"}} 
		and version = {{arg "version"}}
		and language = {{arg "language"}}
	`),
	get: data.NewQuery(`
		select 	d.id,
			d.created,
			d.creator,
			c.language,
			c.title,
			c.content,
			c.version,
			c.updated,
			c.created,
			c.creator,
			c.updater,
			t.tag,
			t.language,
			t.type,
			t.stem,
			g.group_id,
			g.can_publish
		from 	documents d
			inner join document_contents c on d.id = c.document_id
			left outer join document_groups g on d.id = g.document_id
			left outer join document_tags t 
				on d.id = t.document_id
				and c.language = t.language
		where 	d.id = {{arg "id"}}
		and 	c.language = {{arg "language"}}
	`),
	getDraft: data.NewQuery(`
		select 	d.id,
			d.document_id,
			d.language,
			d.title,
			d.content,
			d.version,
			d.updated,
			d.created,
			d.creator,
			d.updater,
			t.tag,
			t.language,
			t.type,
			t.stem
		from 	document_drafts d
			left outer join document_draft_tags t 
				on d.id = t.draft_id
				and d.language = t.language
		where 	d.id = {{arg "id"}}
		and 	d.language = {{arg "language"}}
	`),

	deleteGroup: data.NewQuery(`
		delete from document_groups 
		where 	document_id = {{arg "document_id"}}
		and 	group_id = {{arg "group_id"}}
	`),
	deleteTags: data.NewQuery(`
		delete from document_tags
		where document_id = {{arg "document_id"}}
		and language = {{arg "language"}}
	`),
	deleteDraftTags: data.NewQuery(`
		delete from document_draft_tags
		where draft_id = {{arg "draft_id"}}
		and language = {{arg "language"}}
	`),
	deleteDraft: data.NewQuery(`
		delete from document_drafts
		where 	id = {{arg "id"}}
		and 	language = {{arg "language"}}
	`),
	groupExists: data.NewQuery(`
		select count(*)
		from document_groups
		where document_id = {{arg "document_id"}}
		and group_id = {{arg "group_id"}}
	`),
	updateGroup: data.NewQuery(`
		update document_groups
		set can_publish = {{arg "can_publish"}}
		where document_id = {{arg "document_id"}}
		and group_id = {{arg "group_id"}}
	`),
	canPublish: data.NewQuery(`
		select 	count(*)
		from 	documents d	
			left outer join document_groups g 
				on g.document_id = d.id
				and g.can_publish = {{TRUE}}
			left outer join group_users u on g.group_id = u.group_id
		where	d.id = {{arg "document_id"}}
		and	(
			u.user_id = {{arg "user_id"}}
		or	d.creator = {{arg "creator"}}
		)
	`),
}
