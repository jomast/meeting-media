package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func WeekOf(date time.Time) time.Time {
	return RelativeDay(date, time.Monday)
}

func RelativeDay(date time.Time, toDOW time.Weekday) (out time.Time) {
	out = date
	for out.Weekday() != time.Monday {
		out = out.AddDate(0, 0, -1)
	}
	for out.Weekday() != toDOW {
		out = out.AddDate(0, 0, 1)
	}
	return
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.Mkdir(dir, fs.FileMode(0777)); err != nil {
			return err
		}
	}
	return nil
}
