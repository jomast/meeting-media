package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
)

func (c *Config) fetchMeetingStuff(m string) (err error) {
	logrus.Debug("fetchMeetingStuff()")

	if c.PurgeSaveDir {
		logrus.Info("Deleting all files in " + c.SaveLocation)
		if err := RemoveContents(c.SaveLocation); err != nil {
			logrus.Warn(err)
		}
	}

	if c.AutoFetchMeetingData {
		logrus.Info("Auto-Fetching!")
		var data MeetingData
		var err error
		switch m {
		case WM:
			data, err = c.getWMData()
			c.Songs = []string{
				c.Songs[0],
				data.Songs[0],
				data.Songs[1],
			}
		case MM:
			data, err = c.getMMData()
			c.Songs = data.Songs
			c.Videos = data.Videos
		}
		if err != nil {
			return err
		}

		c.Pictures = data.Pictures
	}

	for _, song := range c.Songs {
		if err := c.downloadSong(song); err != nil {
			return err
		}
	}

	if c.FetchOtherMedia {
		for i, video := range c.Videos {
			name, err := c.downloadVideo(&video)
			if err != nil {
				return err
			}
			c.Videos[i].Name = name
		}

		for _, picture := range c.Pictures {
			logrus.Infof("saving picture %s", picture.Name)
			file := filepath.Join(c.SaveLocation, picture.Name)
			if ioutil.WriteFile(file, picture.Payload, 0644) != nil {
				return errors.New("error writing data to " + file)
			}
		}
	}

	if c.CreatePlaylist {
		return c.createPlaylist()
	}

	return
}

func (c *Config) createPlaylist() error {
	logrus.Info("creating playlist")

	sort.Slice(c.Pictures, func(i, j int) bool {
		return c.Pictures[i].Name < c.Pictures[j].Name
	})

	file := filepath.Join(c.SaveLocation, "/playlist.m3u")
	body := ""
	for _, s := range c.Songs {
		body += s + ".mp4\n"
	}
	for _, v := range c.Videos {
		body += v.Name + "\n"
	}
	for _, p := range c.Pictures {
		body += p.Name + "\n"
	}

	if err := ioutil.WriteFile(file, []byte(body), 0644); err != nil {
		return err
	}

	return nil
}

func (c *Config) downloadSong(num string) (err error) {
	logrus.Info("downloading song " + num)
	var res int
	switch c.Resolution {
	case RES240:
		res = 0
	case RES360:
		res = 1
	case RES480:
		res = 2
	case RES720:
		res = 3
	default:
		res = 0
	}

	path := fmt.Sprintf("%s/%s.mp4", c.SaveLocation, num)

	songInfo, err := c.getSongInfo(num)
	if err != nil {
		return
	}
	err = c.downloadSongMedia(songInfo, res, path)
	if err != nil {
		return
	}

	return nil
}

func (c *Config) downloadVideo(v *video) (name string, err error) {
	var res int
	switch c.Resolution {
	case RES240:
		res = 0
	case RES360:
		res = 1
	case RES480:
		res = 2
	case RES720:
		res = 3
	default:
		res = 0
	}

	var path, url string
	var filesize int
	if v.IssueTagNumber == 0 {
		vidInfo, err := c.getMediaVideoInfo(v)
		if err != nil {
			return "", err
		}
		path = fmt.Sprintf("%s/%s", c.SaveLocation, filepath.Base(vidInfo.Files[c.Language].MP4[res].File.URL))
		url = vidInfo.Files[c.Language].MP4[res].File.URL
		filesize = vidInfo.Files[c.Language].MP4[res].Filesize

	} else if v.IssueTagNumber != 0 {
		vidInfo, err := c.getPubVideoInfo(v)
		if err != nil {
			return "", err
		}

		for i, v := range vidInfo.Media[0].Files {
			if v.Label == c.Resolution && !v.Subtitled {
				res = i
				break
			}
		}

		path = fmt.Sprintf("%s/%s", c.SaveLocation, filepath.Base(vidInfo.Media[0].Files[res].Progressivedownloadurl))
		url = vidInfo.Media[0].Files[res].Progressivedownloadurl
		filesize = vidInfo.Media[0].Files[res].Filesize
	}

	name = filepath.Base(path)

	logrus.Infof("downloading video: %s", name)

	err = c.downloadVideoMedia(url, filesize, path)
	if err != nil {
		return
	}

	return
}

func (c *Config) downloadSongMedia(songInfo *mediaInfo, vidKey int, file string) error {
	url := songInfo.Files[c.Language].MP4[vidKey].File.URL
	if *c.DebugMode {
		logrus.Debug("Mock downloadSongMedia:", url)
	} else {
		logrus.Debug("downloading media " + url)

		if *c.DebugMode {
			logrus.Debug("Mock Download SongMedia:", url)
		} else {
			resp, err := c.HttpClient.Get(url)
			if err != nil {
				return errors.New("failed to download " + url)
			}

			c.Progress.ProgressBar.SetValue(0)
			c.Progress.Total = 0
			c.Progress.Title = filepath.Base(url)
			c.Progress.ProgressBar.Max = float64(songInfo.Files[c.Language].MP4[vidKey].Filesize)

			data := io.TeeReader(resp.Body, c.Progress)

			body, err := ioutil.ReadAll(data)
			if err != nil {
				return errors.New("error reading data from " + url)
			}

			logrus.Debug("writing to  " + file)
			if ioutil.WriteFile(file, body, 0644) != nil {
				return errors.New("error writing data to " + file)
			}
		}
	}
	return nil
}

func (c *Config) downloadVideoMedia(url string, filesize int, file string) error {
	if *c.DebugMode {
		logrus.Debug("Mock downloadVideoMedia:", url)
	} else {
		logrus.Debug("downloading media " + url)
		resp, err := c.HttpClient.Get(url)
		if err != nil {
			return errors.New("failed to download " + url)
		}

		c.Progress.ProgressBar.SetValue(0)
		c.Progress.Total = 0
		c.Progress.Title = filepath.Base(url)
		c.Progress.ProgressBar.Max = float64(filesize)

		data := io.TeeReader(resp.Body, c.Progress)

		body, err := ioutil.ReadAll(data)
		if err != nil {
			return errors.New("error reading data from " + url)
		}

		logrus.Debug("writing to  " + file)
		if ioutil.WriteFile(file, body, 0644) != nil {
			return errors.New("error writing data to " + file)
		}
	}
	return nil
}

func (c *Config) getSongInfo(num string) (*mediaInfo, error) {
	logrus.Debug("fetching info for song number " + num)
	resp, err := c.HttpClient.Get(fmt.Sprintf("https://pubmedia.jw-api.org/GETPUBMEDIALINKS?output=json&pub=sjjm&fileformat=mp4&alllangs=0&track=%s&langwritten=%s&txtCMSLang=%s", num, c.Language, c.Language))
	if err != nil {
		return nil, errors.New("failed to get media info for song #" + num)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("error reading info for song #" + num)
	}

	info := new(mediaInfo)
	err = json.Unmarshal(body, info)

	logrus.Debugf("fetched #%v: %#v", num, info)
	return info, err
}

func (c *Config) getMediaVideoInfo(v *video) (*mediaInfo, error) {
	logrus.Debugf("fetching info for video: %#v ", v)
	variable := ""
	if v.MepsDocumentID.Valid {
		variable = fmt.Sprintf("&docid=%v", v.MepsDocumentID.Int64)
	} else {
		variable = fmt.Sprintf("&pub=%s", v.KeySymbol.String)
	}
	url := fmt.Sprintf("https://pubmedia.jw-api.org/GETPUBMEDIALINKS?%s&output=json&fileformat=mp4&alllangs=0&track=%v&langwritten=%s&txtCMSLang=%s", variable, v.Track.Int64, c.Language, c.Language)

	logrus.Debug("getMediaVideoInfo() url:", url)
	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get media info for video: %#v", v)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading info for video: %#v", v)
	}

	info := new(mediaInfo)
	err = json.Unmarshal(body, info)

	logrus.Debug("getMediaVideoInfo() info:", info)
	return info, err
}

func (c *Config) getPubVideoInfo(v *video) (*PubVideo, error) {
	logrus.Debugf("fetching info for video: %#v ", v)
	url := fmt.Sprintf("https://b.jw-cdn.org/apis/mediator/v1/media-items/%s/pub-%s_%v_%v_VIDEO", c.Language, v.KeySymbol.String, v.IssueTagNumber/100, v.Track.Int64)
	logrus.Debug("getVideoInfo() url:", url)

	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get media info for video: %#v", v)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading info for video: %#v", v)
	}

	info := new(PubVideo)
	err = json.Unmarshal(body, info)

	logrus.Debug("getVideoInfo() info:", info)
	return info, err
}

func (wc *progress) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += int64(n)
	wc.ProgressBar.SetValue(float64(wc.Total))
	return n, nil
}
