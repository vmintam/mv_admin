package main

import (
	"fmt"
	//	"time"
	pb "muvik/muvik_admin/protos/comment"

	"github.com/garyburd/redigo/redis"
)

const (
	COMMENT_PREFIX = "c::"
	REPLY_COMMENT  = `c::%s::r`
)

func gcommentDetail(commentID string) (commentDetail map[string]string, err error) {
	conn := comment_pool.Get()
	defer conn.Close()
	commentDetail, err = redis.StringMap(conn.Do("HGETALL", COMMENT_PREFIX+commentID))
	if err == redis.ErrNil {
		return commentDetail, nil
	}
	return
}

func scommentDetail(commentID string, fields map[string]string) (err error) {
	conn := comment_pool.Get()
	defer conn.Close()
	args := convertMaptoArray(commentID, fields)
	_, err = conn.Do("HMSET", args...)
	return
}

func dcommentDetail(commentID string, fields []string) (err error) {
	conn := comment_pool.Get()
	defer conn.Close()
	if fields == nil {
		_, err = conn.Do("HCLEAR")
		return err
	} else {
		args := convertVL(commentID, fields)
		_, err = conn.Do("HDEL", args)
		return err
	}

}

func countReplyComment(commentID string) (total int, err error) {
	conn := comment_pool.Get()
	defer conn.Close()
	total, err = redis.Int(conn.Do("ZCARD", fmt.Sprintf(REPLY_COMMENT, commentID)))
	if err == redis.ErrNil {
		return 0, nil
	}
	return
}

func glistReplyComment(commentID string) (reply map[string]string, err error) {
	conn := comment_pool.Get()
	defer conn.Close()
	reply, err = redis.StringMap(conn.Do("ZRANGE", fmt.Sprintf(REPLY_COMMENT, commentID), "0", "-1", "withscores"))
	if err == redis.ErrNil {
		return reply, nil
	}
	return
}

func addtolistReplyComment(commentID string, replies []*pb.ReplyComment) (err error) {
	conn := comment_pool.Get()
	defer conn.Close()
	for _, r := range replies {
		_, err = conn.Do("ZADD", fmt.Sprintf(REPLY_COMMENT, commentID), r.Score, r.ReplyID)
		if err != nil {
			return err
		}
	}
	return nil
}

func dlistReplyComment(commentID string, replies []string) (err error) {
	conn := comment_pool.Get()
	defer conn.Close()
	args := convertVL(commentID, replies)
	_, err = conn.Do("ZREM", args)
	return nil
}

func convertVL(first string, input []string) (output []interface{}) {
	output = append(output, first)
	for _, v := range input {
		output = append(output, v, "1")
	}
	return
}

func convertMaptoArray(first string, input map[string]string) (output []interface{}) {
	output = append(output, first)
	for k, v := range input {
		output = append(output, k, v)
	}
	return
}

func revMaptoArray(first string, input map[string]string) (output []interface{}) {
	output = append(output, first)
	for k, v := range input {
		output = append(output, v, k)
	}
	return
}
