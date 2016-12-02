package main

import (
	"fmt"
	"time"

	pb "muvik/muvik_admin/protos/video"

	"github.com/garyburd/redigo/redis"
)

const (
	VIDEO_PREFIX     = "v::"
	VIDEO_COVER      = "image"
	VIDEO_TIMESTAMP  = "timestamp"
	VIDEO_TOTAL_VIEW = "total_view"
	COMMENTS_VIDEO   = `v::%s::c`
	VIDEO_PROMOTE    = "v::promote"
)

func convertTimeStamp(ts int64) (string, error) {
	//convert to local time
	datetime := time.Unix(ts, 0)
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err == nil {
		datetime = datetime.In(location)
		return fmt.Sprint(datetime), nil
	}
	return "", err
}

func getVideoDetail(videoID string) (video map[string]string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	// select in redis 9502 -> get total_view video
	video, err = redis.StringMap(conn.Do("HGETALL", VIDEO_PREFIX+videoID))
	return
}

func getCoverImage(videoID string) (cover string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	//get video create time
	created, err := redis.Int64(conn.Do("HGET", VIDEO_PREFIX+videoID, VIDEO_TIMESTAMP))
	if err != nil {
		return "", err
	}
	//convert ts
	datetime := time.Unix(created/1000, 0)
	year := datetime.Year()
	month := int(datetime.Month())
	day := datetime.Day()
	// select in redis 9502 -> get total_view video
	cover, err = redis.String(conn.Do("HGET", VIDEO_PREFIX+videoID, VIDEO_COVER))
	cover = fmt.Sprintf("%d/%d/%d/%s", year, month, day, cover)
	return
}

func getVideoCreated(videoID string) (created int64, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	// select in redis 9502 -> get total_view video
	created, err = redis.Int64(conn.Do("HGET", VIDEO_PREFIX+videoID, VIDEO_TIMESTAMP))
	return
}

func getTotalViews(videoID string) (totalviews int, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	// select in redis 9502 -> get total_view video
	totalviews, err = redis.Int(conn.Do("HGET", VIDEO_PREFIX+videoID, VIDEO_TOTAL_VIEW))
	return
}

func deleteVideo(videoID string, fields map[string]string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	if fields != nil {
		for field_name, _ := range fields {
			_, err = conn.Do("HDEL", VIDEO_PREFIX+videoID, field_name)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		_, err = conn.Do("DEL", VIDEO_PREFIX+videoID)
		if err != nil {
			return err
		}
		return nil
	}
	// select in redis 9502 -> get total_view video

}

func updateVideo(videoID string, fields map[string]string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	if fields != nil {
		return fmt.Errorf("field not nil")
	} else {
		for k, v := range fields {
			_, err = conn.Do("HMSET", VIDEO_PREFIX+videoID, k, v)
			if err != nil {
				return err
			}
		}
		return nil
	}

}

func GetListComments(videoID string) (comments map[string]string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	comments, err = redis.StringMap(conn.Do("ZRANGE", fmt.Sprintf(COMMENTS_VIDEO, videoID), "0", "-1", "withscores"))
	if err == redis.ErrNil {
		return comments, nil
	}
	return
}

func AddComments(videoID string, css []*pb.CommentsVideo) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	for _, cs := range css {
		_, err = conn.Do("ZADD", fmt.Sprintf(COMMENTS_VIDEO, videoID), &cs.Score, &cs.CommentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteComments(videoID string, commentid []string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	for _, c := range commentid {
		_, err = conn.Do("ZREM", fmt.Sprintf(COMMENTS_VIDEO, videoID), c)
		if err != nil {
			return err
		}
	}
	return nil
}

func GPromoteVideo() (videoID string, err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	videoID, err = redis.String(conn.Do("GET", VIDEO_PROMOTE))
	return
}

func SPromoteVideo(videoID string) (err error) {
	conn := video_pool_info.Get()
	defer conn.Close()
	_, err = conn.Do("SET", VIDEO_PROMOTE, videoID)
	return
}
