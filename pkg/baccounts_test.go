package baccounts

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
		t.Fatal("Cannot decode: ", e)
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

	f, err := ioutil.TempFile("/tmp", "baccount-test.json")
	if err != nil {
		t.Error("Can't create temporary file")
	}

	b := &Baccount{Profiles: profiles}

	_ = f.Truncate(0)
	s, _ := b.toJson()
	_, _ = f.Write([]byte(s))
	_ = f.Sync()
	//log.Println("ok")
	_, _ = f.Seek(0, 0)
	f.Close()

	b2, err := LoadKeysFromJson(s)
	if err != nil {
		t.Error("Can't read test json file", err)
	}
	profile := b2.Profiles[0]

	if profile.Name != "me@mac.com" {
		t.Error("Name")
	}
	if profile.Sites["a.com"].Url != "a" {

		t.Error("a.com", profile.Sites["a"])
	}
	//var as []Profile
	// b, _ := ioutil.ReadAll(f)
	//log.Println(string(b))
	// e := json.Unmarshal(b, &as)
	//if e != nil {
	//t.Fatal(">>>")
	//}
	//log.Printf("read: %v", as)
}
