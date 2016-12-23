package main

import (
	"fmt"
	//	"time"
	pb "muvik/muvik_admin/protos/user"

	"github.com/garyburd/redigo/redis"
)

const (
	//hash detail
	USER_DETAIL      = `u::%s`
	USER_FBID_USERID = `u:fb`
	USER_FB_FRIEND   = `u::fb::friend`
	USER_FRIEND      = `friend::u::%s` //friend::u::{userid}
	//list
	LIST_VIDEOS_OF_USER           = `u::%s::v`         //u::{user_id}::v
	LIST_VIDEOS_SHARED_BY_USER    = `u:%s:s`           //u:{user_id}:s
	LIST_AUDIO_INTERESTED_BY_USER = `u::%s::a	`        //u::{user_id}::a
	LIST_FB_FRIEND_OF_USER        = `u::%s::fb_friend` //u::{user_id}::fb_friend
	LIST_FANS_OF_USER             = `u::%s::f`         //u::{user_id}::f
	LIST_IDOLS_OF_USER            = `u::%s::i`         //u::{user_id}::i
)

func gListOne(requestID string, listtype pb.UserListType, member string) (result string, err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.UserListType_listVideoOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_OF_USER, requestID)
	case pb.UserListType_listVideoSharedOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_SHARED_BY_USER, requestID)
	case pb.UserListType_listFBFriendOfUser:
		key = fmt.Sprintf(LIST_FB_FRIEND_OF_USER, requestID)
	case pb.UserListType_listAudioInterestOfUser:
		key = fmt.Sprintf(LIST_AUDIO_INTERESTED_BY_USER, requestID)
	case pb.UserListType_listFansOfUser:
		key = fmt.Sprintf(LIST_FANS_OF_USER, requestID)
	case pb.UserListType_listIdolOfUser:
		key = fmt.Sprintf(LIST_IDOLS_OF_USER, requestID)
	default:
		return
	}
	result, err = redis.String(conn.Do("ZSCORE", key, member))
	if err == redis.ErrNil {
		return "", nil
	}
	return
}

func gList(requestID string, listtype pb.UserListType) (list map[string]string, err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.UserListType_listVideoOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_OF_USER, requestID)
	case pb.UserListType_listVideoSharedOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_SHARED_BY_USER, requestID)
	case pb.UserListType_listFBFriendOfUser:
		key = fmt.Sprintf(LIST_FB_FRIEND_OF_USER, requestID)
	case pb.UserListType_listAudioInterestOfUser:
		key = fmt.Sprintf(LIST_AUDIO_INTERESTED_BY_USER, requestID)
	case pb.UserListType_listFansOfUser:
		key = fmt.Sprintf(LIST_FANS_OF_USER, requestID)
	case pb.UserListType_listIdolOfUser:
		key = fmt.Sprintf(LIST_IDOLS_OF_USER, requestID)
	default:
		return
	}
	list, err = redis.StringMap(conn.Do("ZRANGE", key, "0", "-1", "withscores"))
	if err == redis.ErrNil {
		return list, nil
	}
	return
}

func aList(requestID string, listtype pb.UserListType, member_scores map[string]string) (err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.UserListType_listVideoOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_OF_USER, requestID)
	case pb.UserListType_listVideoSharedOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_SHARED_BY_USER, requestID)
	case pb.UserListType_listFBFriendOfUser:
		key = fmt.Sprintf(LIST_FB_FRIEND_OF_USER, requestID)
	case pb.UserListType_listAudioInterestOfUser:
		key = fmt.Sprintf(LIST_AUDIO_INTERESTED_BY_USER, requestID)
	case pb.UserListType_listFansOfUser:
		key = fmt.Sprintf(LIST_FANS_OF_USER, requestID)
	case pb.UserListType_listIdolOfUser:
		key = fmt.Sprintf(LIST_IDOLS_OF_USER, requestID)
	default:
		return
	}
	args := revMaptoArray(key, member_scores)
	_, err = conn.Do("ZADD", args...)
	return
}

func rList(requestID string, listtype pb.UserListType, member_scores map[string]string) (err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.UserListType_listVideoOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_OF_USER, requestID)
	case pb.UserListType_listVideoSharedOfUser:
		key = fmt.Sprintf(LIST_VIDEOS_SHARED_BY_USER, requestID)
	case pb.UserListType_listFBFriendOfUser:
		key = fmt.Sprintf(LIST_FB_FRIEND_OF_USER, requestID)
	case pb.UserListType_listAudioInterestOfUser:
		key = fmt.Sprintf(LIST_AUDIO_INTERESTED_BY_USER, requestID)
	case pb.UserListType_listFansOfUser:
		key = fmt.Sprintf(LIST_FANS_OF_USER, requestID)
	case pb.UserListType_listIdolOfUser:
		key = fmt.Sprintf(LIST_IDOLS_OF_USER, requestID)
	default:
		return
	}
	if member_scores == nil {
		_, err = conn.Do("ZCLEAR", key)
		if err != nil {
			return
		}
		return nil
	} else {
		for member, _ := range member_scores {
			_, err = conn.Do("ZREM", key, member)
			if err != nil {
				return err
			}
		}
		return nil
	}

}

func gDetailOne(requestID string, kind pb.UserKind, field string) (result string, err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch kind {
	case pb.UserKind_detail:
		key = fmt.Sprintf(USER_DETAIL, requestID)
	case pb.UserKind_fbid_userid:
		key = USER_FBID_USERID
	case pb.UserKind_fb_friend:
		key = USER_FB_FRIEND
	case pb.UserKind_fb_userid:
		key = fmt.Sprintf(USER_FRIEND, requestID)
	default:
		break
	}
	result, err = redis.String(conn.Do("HGET", key, field))
	if err == redis.ErrNil {
		return "", nil
	}
	return
}

func gDetail(requestID string, kind pb.UserKind) (Detail map[string]string, err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch kind {
	case pb.UserKind_detail:
		key = fmt.Sprintf(USER_DETAIL, requestID)
	case pb.UserKind_fbid_userid:
		key = USER_FBID_USERID
	case pb.UserKind_fb_friend:
		key = USER_FB_FRIEND
	case pb.UserKind_fb_userid:
		key = fmt.Sprintf(USER_FRIEND, requestID)
	default:
		break
	}
	Detail, err = redis.StringMap(conn.Do("HGETALL", key))
	if err == redis.ErrNil {
		return Detail, nil
	}
	return
}

func sDetail(requestID string, kind pb.UserKind, fields map[string]string) (err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch kind {
	case pb.UserKind_detail:
		key = fmt.Sprintf(USER_DETAIL, requestID)
	case pb.UserKind_fbid_userid:
		key = USER_FBID_USERID
	case pb.UserKind_fb_friend:
		key = USER_FB_FRIEND
	case pb.UserKind_fb_userid:
		key = fmt.Sprintf(USER_FRIEND, requestID)
	default:
		break
	}
	if fields == nil {
		return nil
	} else {
		args := convertMaptoArray(key, fields)
		_, err = conn.Do("HMSET", args...)
		if err != nil {
			return
		}
		return nil
	}
}

func dDetail(requestID string, kind pb.UserKind, fields map[string]string) (err error) {
	conn := user_pool.Get()
	defer conn.Close()
	key := ""
	switch kind {
	case pb.UserKind_detail:
		key = fmt.Sprintf(USER_DETAIL, requestID)
	case pb.UserKind_fbid_userid:
		key = USER_FBID_USERID
	case pb.UserKind_fb_friend:
		key = USER_FB_FRIEND
	case pb.UserKind_fb_userid:
		key = fmt.Sprintf(USER_FRIEND, requestID)
	default:
		break
	}
	if fields == nil {
		_, err = conn.Do("HCLEAR", key)
		if err != nil {
			return
		}
		return nil
	} else {
		args := convertMaptoArray(key, fields)
		_, err = conn.Do("HDEL", args...)
		if err != nil {
			return
		}
		return nil
	}
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
