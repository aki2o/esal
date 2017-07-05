package action

import (
	"flag"
	"strconv"
	"bufio"
	"os"
	"unicode/utf8"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/upamune/go-esa/esa"
)

type save struct {}

func init() {
	processors["save"] = &save{}
}

func (self *save) SetOption(flagset *flag.FlagSet) {
}

func (self *save) Do(args []string) error {
	post_number			:= args[0]
	scan_byte_size, err := strconv.Atoi(args[1])
	scanned_byte_size	:= 0
	post_scanner		:= bufio.NewScanner(os.Stdin)
	post_string			:= ""

	if err != nil { return err }
	
	log.WithFields(log.Fields{ "number": post_number, "size": scan_byte_size }).Debug("start scan")
	for scanned_byte_size < scan_byte_size && post_scanner.Scan() {
		post_string += post_scanner.Text()+"\r\n"
		scanned_byte_size = utf8.RuneCountInString(post_string)
		
		log.WithFields(log.Fields{ "scanned_size": scanned_byte_size }).Debug("scanned")
	}

	log.WithFields(log.Fields{ "number": post_number, "size": scan_byte_size }).Debug("start update post")
	var post esa.Post
	bytes := LoadPostData(Context.Cwd, post_number, "json")
	if err = json.Unmarshal(bytes, &post); err != nil { return err }
	
	post.BodyMd = post_string
	
	post_number_i, _ := strconv.Atoi(post_number)
	res, err := Context.Client.Post.Update(Context.Team, post_number_i, post)
	if err != nil { return err }
	SavePost(res)

	return nil
}
