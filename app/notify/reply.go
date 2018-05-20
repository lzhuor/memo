package notify

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/cache"
	"github.com/memocash/memo/app/db"
	"time"
)

type ReplyNotification struct {
	Post         *db.MemoPost
	Parent       *db.MemoPost
	Notification *db.Notification
	Name         string
}

func (n ReplyNotification) GetName() string {
	return n.Name
}

func (n ReplyNotification) GetAddressString() string {
	return n.Post.GetAddressString()
}

func (n ReplyNotification) GetPostHashString() string {
	return n.Post.GetTransactionHashString()
}

func (n ReplyNotification) GetMessage() string {
	return n.Post.GetMessage()
}

func (n ReplyNotification) GetLink() string {
	hash, err := chainhash.NewHash(n.Post.TxHash)
	if err != nil {
		jerr.Get("error getting like notification tx hash", err).Print()
		return ""
	}
	return fmt.Sprintf("post/%s", hash.String())
}

func (n ReplyNotification) GetTime() time.Time {
	if n.Post.Block != nil {
		return n.Post.Block.Timestamp
	} else {
		return n.Post.CreatedAt
	}
}

func AddReplyNotification(reply *db.MemoPost, updateCache bool) error {
	parent, err := db.GetMemoPost(reply.ParentTxHash)
	if err != nil {
		return jerr.Get("error getting parent post", err)
	}
	userId, err := db.GetUserIdFromPkHash(parent.PkHash)
	if err != nil {
		if db.IsRecordNotFoundError(err) {
			// Don't add notifications for external users, not an error though
			return nil
		}
		return jerr.Get("error getting user id from pk hash", err)
	}
	_, err = db.AddNotification(parent.PkHash, reply.TxHash, db.NotificationTypeReply)
	if err != nil {
		return jerr.Get("error adding notification", err)
	}
	if updateCache {
		_, err = cache.GetAndSetUnreadNotificationCount(userId)
		if err != nil {
			return jerr.Get("error setting notification unread count", err)
		}
	}
	return nil
}