package main

import (
	"net"
	"fmt"
	"bufio"
	"strings"
	"flag"
	"strconv"
	"encoding/xml"
)


const ERRORTYPE_ServerError = 6
const ERRORTYPE_SMTPCMDError = 5
const ERRORTYPE_HelloError = 1
const ERRORTYPE_HelloDomainError = 2
const ERRORTYPE_SenderError = 3
const ERRORTYPE_RecipientRejected = 4
const ALL_OK = 7


var (
	ESAs = flag.String("s", "", "Cisco ESA server IP or URI, separated by commas")
	mailfrom = flag.String("f", "", "Mail from")
	rcptto = flag.String("t", "", "mail to")
	debuggmode = flag.Bool("d",false,"Debbug mode, det to True if You want get additional information")
)

type result struct {
	Channel string `xml:"channel"`
	Value string `xml:"value"`
	Lookup string `xml:"ValueLookup"`
}

type prtgbody struct {
	XMLName   xml.Name `xml:"prtg"`
	TextField string   `xml:"text"`
	Res       []result `xml:"result"`
}

func TestESA(ESA string,MailFrom string,MailTo string)int{
	conn, errconn := net.Dial("tcp", ESA+":25")
	if errconn != nil{
		return ERRORTYPE_ServerError
	}
	defer conn.Close()
	retval := ERRORTYPE_ServerError

	spdomain := strings.Split(MailFrom,"@")

	message,err := bufio.NewReader(conn).ReadString('\n')
	if err != nil{
		fmt.Println(err)
	}
	if *debuggmode  {
		fmt.Println(message)
	}
	mesfromservspited := strings.Split(message," ")
	ReturnCode,_ := strconv.Atoi(mesfromservspited[0])
	if ReturnCode != 220{
		retval = ERRORTYPE_HelloError
		return retval
	}

	fmt.Fprintf(conn, "HELO "+spdomain[1]+ "\n")
	message,err = bufio.NewReader(conn).ReadString('\n')
	if *debuggmode  {
		fmt.Println(message)
	}
	mesfromservspited = strings.Split(message," ")
	ReturnCode,_ = strconv.Atoi(mesfromservspited[0])
	if ReturnCode != 250{
		retval = ERRORTYPE_HelloDomainError
		return retval
	}

	fmt.Fprintf(conn, "MAIL FROM:"+MailFrom+ "\n")
	message,err = bufio.NewReader(conn).ReadString('\n')
	if *debuggmode  {
		fmt.Println(message)
	}
	mesfromservspited = strings.Split(message," ")
	ReturnCode,_ = strconv.Atoi(mesfromservspited[0])
	if ReturnCode != 250{
		retval = ERRORTYPE_SenderError
		return retval
	}

	fmt.Fprintf(conn, "RCPT TO:"+MailTo+ "\n")
	message,err = bufio.NewReader(conn).ReadString('\n')
	if *debuggmode  {
		fmt.Println(message)
	}
	mesfromservspited = strings.Split(message," ")
	ReturnCode,_ = strconv.Atoi(mesfromservspited[0])
	if ReturnCode != 250{
		retval = ERRORTYPE_RecipientRejected
		return retval
	}

	fmt.Fprintf(conn, "QUIT" + "\n")
	message,err = bufio.NewReader(conn).ReadString('\n')
	if *debuggmode  {
		fmt.Println(message)
	}
	mesfromservspited = strings.Split(message," ")
	ReturnCode,_ = strconv.Atoi(mesfromservspited[0])
	if ReturnCode != 221{
		retval = ERRORTYPE_SMTPCMDError
		return retval
	}
	return ALL_OK
}

func main() {
	var rd1 []result
	flag.Parse()
	ESAsList := strings.Split(*ESAs,",")
	for _,ESA := range ESAsList{
		val := TestESA(ESA,*mailfrom,*rcptto)
		rd1 = append(rd1,result{Channel:ESA,Value:strconv.Itoa(val),Lookup:"esasmtplookup"})
	}
	mt1 := &prtgbody{TextField:"",Res: rd1}
	bolB, _ := xml.Marshal(mt1)
	fmt.Println(string(bolB))
}
