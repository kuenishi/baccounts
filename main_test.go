package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"
)

func TestBasic(t *testing.T) {
	profiles := make([]Profile, 0, 16)

	a := NewProfile("me@mac.com", true)
	a.AddSite("a.com", "a", "b", "c", "me@mac.com")
	a.AddSite("b.com", "s", "b", "x", "me@mac.com")

	profiles = append(profiles, *a)

	_, e := json.Marshal(profiles)
	if e != nil {
		t.Fatal("Cannot decode: %v", e)
	}
	log.Printf("len=%v, cap=%v %v", len(profiles), cap(profiles), profiles)
	//log.Println("Encoded: ", string(b))
}

func TestSave(t *testing.T) {
	p := NewProfile("me@mac.com", true)
	p.AddSite("a.com", "a", "b", "c", "me@ma.com")
	p.AddSite("b.com", "s", "b", "x", "me@mac.com")

	profiles := make([]*Profile, 0, 16)
	profiles = append(profiles, p)

	f, e2 := ioutil.TempFile("/tmp", "baccount-test.json")
	if e2 != nil {
		t.Error("Can't create temporary file")
	}
	_ = f.Truncate(0)
	s, _ := toJson(profiles)
	_, _ = f.Write([]byte(s))
	_ = f.Sync()
	//log.Println("ok")
	_, _ = f.Seek(0, 0)

	var as []Profile
	b, _ := ioutil.ReadAll(f)
	//log.Println(string(b))
	e := json.Unmarshal(b, &as)
	if e != nil {
		t.Fatal(">>>")
	}
	//log.Printf("read: %v", as)
}
