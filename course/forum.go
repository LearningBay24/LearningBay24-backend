package course

import (
	"context"
	"database/sql"

	"learningbay24.de/backend/models"

	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Create a new forum_entry (= forum post) in a specific forum.
// Set `inReplyTo` parameter to `0` if the post is a parent post and doesn't reply to anyone.
// Returns the newly created forum_entry id, or an error when inserting the new forum_entry fails.
func CreateForumEntry(db *sql.DB, forumId int, userId int, content string, inReplyTo int) (int, error) {
	flog := log.WithFields(log.Fields{
		"context": "CreateForumEntry",
	})

	reply := null.IntFrom(inReplyTo)
	if inReplyTo == 0 {
		reply = null.Int{}
	}

	forum_entry := models.ForumEntry{ForumID: forumId, Content: content, AuthorID: userId, InReplyTo: reply}
	err := forum_entry.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		flog.Errorf("Unable to insert forum_entry: %s", err.Error())
		return 0, err
	}

	return forum_entry.ID, nil
}
