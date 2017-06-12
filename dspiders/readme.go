// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

// dspiders is means Distributed Spider
//
// This is a web page crawl and full text index package.
//
// It's have two parts for dspiders: crawl machine and process machine.
// And they are distributed.
//
// Crawl Machine:
//
// This part as a crawl node can be deployed to everywhere if you want.
// It connect to Process Machine use the nst package.
// It get the url infomation from Process Machine and crawl the page.
// If the page was changed,
// send the page content (the content just in <body> tag, and be removed all html tag) to Process Machine (the type is PageData),
// analysis the links in the page (html <a> tag), push all links to Process Machine (the type is UrlBasic).
// At present, the Crawl Machine don't crawl the media file data (css, js, jpeg, png etc. ).
//
// Process Machine:
//
// This part process all data where from the Crawl Machine. It have many small parts.
//
// NetTransport. It manage the communication between the Crawl Machine and the Process Machine, use the nst package.
//
// UrlCrawlQueue. It manage a queue, which store waiting be crawled urls. When Crawl Machine ask for new url, it will provide from this queue.
//
// PagesProcess. It process the page content data and urls data which from Crawl Machine.
// The urls which need be crawled will send to UrlCrawlQueue.
// The page content will be reprocess and store.
// At last, the reprocessed page content will send to WordsProcess.
//
// WordsProcess. It manage the full text index for page content.
// According to sentences, words, character, it split the page content to slice which recorded the index position.
// Then store the index information for each page, merge/change words context relationship information.
//
// Words Context Relationship Informaion:
//
// The words index store in the dspiders is use the Roles context status. It record the position of a word in the page content, and what word before and after this word in the sentence.
//
// For example the page url is : http://for.example.com/a. The page content is: Today is sunday, We have a funny holiday.
//
// The page content will be process to only text content "today is sunday \n we have a funny holiday". It have no punctuation but only the line break, every line is one sentence.
// Then the text content will be process to a map slice map[uint64][]string. It seems like this: [0]{today, is, sunday},[16]{we, have, a, funny, holiday}.
//
// Next step, the dspiders will store the relationship, for example fide the "is" Role,
// find the "sunday" in the down context relationship, store the link and the position 2 of the word "sunday",
// find the "today" in the up context relationship, store the link and the position 0 of the word "today".
//
// Now we record the relationship which word appear which link, and the word position in the content, and what word before and after this word.
package dspiders
