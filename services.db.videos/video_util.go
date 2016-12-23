package main

import (
	"fmt"
	//	"time"
	pb "muvik/muvik_admin/protos/video"

	"github.com/garyburd/redigo/redis"
)

const (
	VIDEO_DETAIL                    = "v::"
	VIDEO_PROMOTE                   = "v::promote"
	LIST_COMMENTS_VIDEO             = `v::%s::c`
	LIST_COMMENTS_VIDEO_LIKE_WEIGHT = `v::%s::c::l`
)

//video detail

func gVideoDetail(requestID string) (Detail map[string]string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := VIDEO_DETAIL + requestID
	Detail, err = redis.StringMap(conn.Do("HGETALL", key))
	if err == redis.ErrNil {
		return Detail, nil
	}
	return
}

func gVideoOne(requestID string, field string) (One string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := VIDEO_DETAIL + requestID
	One, err = redis.String(conn.Do("HGET", key, field))
	if err == redis.ErrNil {
		return One, nil
	}
	return
}

func sVideoDetail(requestID string, fields map[string]string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := VIDEO_DETAIL + requestID
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

func dVideoDetail(requestID string, fields map[string]string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := VIDEO_DETAIL + requestID
	if fields == nil {
		_, err = conn.Do("HCLEAR", key)
		if err != nil {
			return
		}
		return nil
	} else {
		args := convertMaptoArrayWithOutValue(key, fields)
		_, err = conn.Do("HDEL", args...)
		if err != nil {
			return
		}
		return nil
	}
}

//video promote
func gPromoteVideo() (videoid string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	videoid, err = redis.String(conn.Do("GET", VIDEO_PROMOTE))
	if err == redis.ErrNil {
		return "", nil
	}
	return
}

func sPromoteVideo(videoid string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	if videoid == "" {
		return fmt.Errorf("VideoID not empty")
	}
	_, err = conn.Do("SET", VIDEO_PROMOTE, videoid)
	if err == redis.ErrNil {
		return nil
	}
	return err
}

func gListOne(requestID string, listtype pb.ListType, member string) (score string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.ListType_Comment:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO, requestID)
	case pb.ListType_CommentWithLikeWeight:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO_LIKE_WEIGHT, requestID)
	default:
		break
	}
	score, err = redis.String(conn.Do("ZSCORE", key, member))
	if err == redis.ErrNil {
		return "", nil
	}
	return
}

func gList(requestID string, listtype pb.ListType) (list map[string]string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.ListType_Comment:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO, requestID)
	case pb.ListType_CommentWithLikeWeight:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO_LIKE_WEIGHT, requestID)
	default:
		break
	}
	list, err = redis.StringMap(conn.Do("ZRANGE", key, "0", "-1", "withscores"))
	if err == redis.ErrNil {
		return list, nil
	}
	return
}

func aList(requestID string, listtype pb.ListType, member_scores map[string]string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.ListType_Comment:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO, requestID)
	case pb.ListType_CommentWithLikeWeight:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO_LIKE_WEIGHT, requestID)
	default:
		break
	}
	args := revMaptoArray(key, member_scores)
	_, err = conn.Do("ZADD", args...)
	return
}

func rList(requestID string, listtype pb.ListType, member_scores map[string]string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.ListType_Comment:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO, requestID)
	case pb.ListType_CommentWithLikeWeight:
		key = fmt.Sprintf(LIST_COMMENTS_VIDEO_LIKE_WEIGHT, requestID)
	default:
		break
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

func convertMaptoArrayWithOutValue(first string, input map[string]string) (output []interface{}) {
	output = append(output, first)
	for k, _ := range input {
		output = append(output, k)
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
