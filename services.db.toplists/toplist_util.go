package main

import (
	"fmt"
	//	"time"
	pb "muvik/muvik_admin/protos/toplist"

	"github.com/garyburd/redigo/redis"
)

const (
//list

)

func gList(requestID string, listtype pb.AudioListType) (list map[string]string, err error) {
	conn := audio_pool.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.AudioListType_listAudioByHashTag:
		key = fmt.Sprintf(LIST_AUDIO_BY_HASHTAG, requestID)
	case pb.AudioListType_listAudioEditorChoice:
		key = LIST_AUDIO_EDITOR_CHOICE
	case pb.AudioListType_listAudioSuggestion:
		key = LIST_AUDIO_SUGGESTION
	case pb.AudioListType_listAudioRegular:
		key = fmt.Sprintf(LIST_AUDIO_WITH_RANK, requestID)
	case pb.AudioListType_listTopic:
		key = fmt.Sprintf(LIST_TOPIC_WITH_RANK, requestID)
	case pb.AudioListType_listCategories:
		key = LIST_CATEGORIES_WITH_RANK
	case pb.AudioListType_listAudioInEvent:
		key = fmt.Sprintf(LIST_AUDIO_IN_EVENT, requestID)
	case pb.AudioListType_listEventIDEnd:
		key = LIST_EVENTID_END
	case pb.AudioListType_listUserIDInEvent:
		key = fmt.Sprintf(LIST_USERID_IN_EVENT, requestID)
		arrs, err := redis.Strings(conn.Do("SMEMBERS", key))
		if err == redis.ErrNil {
			return list, nil
		}
		for _, v := range arrs {
			list[v] = "1"
		}
		return list, nil
	default:
		break
	}
	list, err = redis.StringMap(conn.Do("ZRANGE", key, "0", "-1", "withscores"))
	if err == redis.ErrNil {
		return list, nil
	}
	return
}

func aList(requestID string, listtype pb.AudioListType, member_scores map[string]string) (err error) {
	conn := audio_pool.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.AudioListType_listAudioByHashTag:
		key = fmt.Sprintf(LIST_AUDIO_BY_HASHTAG, requestID)
	case pb.AudioListType_listAudioEditorChoice:
		key = LIST_AUDIO_EDITOR_CHOICE
	case pb.AudioListType_listAudioSuggestion:
		key = LIST_AUDIO_SUGGESTION
	case pb.AudioListType_listAudioRegular:
		key = fmt.Sprintf(LIST_AUDIO_WITH_RANK, requestID)
	case pb.AudioListType_listTopic:
		key = fmt.Sprintf(LIST_TOPIC_WITH_RANK, requestID)
	case pb.AudioListType_listCategories:
		key = LIST_CATEGORIES_WITH_RANK
	case pb.AudioListType_listAudioInEvent:
		key = fmt.Sprintf(LIST_AUDIO_IN_EVENT, requestID)
	case pb.AudioListType_listEventIDEnd:
		key = LIST_EVENTID_END
	case pb.AudioListType_listUserIDInEvent:
		key = fmt.Sprintf(LIST_USERID_IN_EVENT, requestID)
		for member, _ := range member_scores {
			_, err := conn.Do("SADD", key, member)
			if err == redis.ErrNil {
				return nil
			}
			return err
		}

	default:
		break
	}
	args := revMaptoArray(key, member_scores)
	_, err = conn.Do("ZADD", args...)
	return
}

func rList(requestID string, listtype pb.AudioListType, member_scores map[string]string) (err error) {
	conn := audio_pool.Get()
	defer conn.Close()
	key := ""
	switch listtype {
	case pb.AudioListType_listAudioByHashTag:
		key = fmt.Sprintf(LIST_AUDIO_BY_HASHTAG, requestID)
	case pb.AudioListType_listAudioEditorChoice:
		key = LIST_AUDIO_EDITOR_CHOICE
	case pb.AudioListType_listAudioSuggestion:
		key = LIST_AUDIO_SUGGESTION
	case pb.AudioListType_listAudioRegular:
		key = fmt.Sprintf(LIST_AUDIO_WITH_RANK, requestID)
	case pb.AudioListType_listTopic:
		key = fmt.Sprintf(LIST_TOPIC_WITH_RANK, requestID)
	case pb.AudioListType_listCategories:
		key = LIST_CATEGORIES_WITH_RANK
	case pb.AudioListType_listAudioInEvent:
		key = fmt.Sprintf(LIST_AUDIO_IN_EVENT, requestID)
	case pb.AudioListType_listEventIDEnd:
		key = LIST_EVENTID_END
	case pb.AudioListType_listUserIDInEvent:
		key = fmt.Sprintf(LIST_USERID_IN_EVENT, requestID)
		for member, _ := range member_scores {
			_, err := conn.Do("SREM", key, member)
			if err == redis.ErrNil {
				return nil
			}
			return err
		}
	default:
		break
	}
	if member_scores == nil {
		_, err = conn.Do("DEL", key)
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

func revMaptoArray(first string, input map[string]string) (output []interface{}) {
	output = append(output, first)
	for k, v := range input {
		output = append(output, v, k)
	}
	return
}
