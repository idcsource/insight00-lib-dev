// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"fmt"
	"strconv"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// The words index process
type WordsIndexProcess struct {
	queue  chan *WordsIndexRequest // the queue
	count  uint                    // the queue's count
	closed bool                    // if closed it will true

	drule *operator.Operator // The DRule2 remote operator

	//sentencedb *reladb.RelaDB // The pages sentence index db

	sentencedb *dbinfo // The pages sentence index db
	worddb     *dbinfo // The word index db
	keyworddb  *dbinfo // The key word index db
}

func NewWordsIndexProcess(sentencedb *operator.Operator, sentencedbname string, worddb *operator.Operator, worddbname string) (w *WordsIndexProcess) {
	w = &WordsIndexProcess{
		queue:      make(chan *WordsIndexRequest, WORDS_PROCESS_QUEUE_CAP),
		count:      0,
		closed:     true,
		sentencedb: &dbinfo{area: sentencedbname, drule: sentencedb},
		worddb:     &dbinfo{area: worddbname, drule: worddb},
	}
	return
}

// Start the processor
func (w *WordsIndexProcess) Start() {
	w.closed = false
	go w.goindex()
}

// Stop the processor
func (w *WordsIndexProcess) Close() {
	w.closed = true
}

// return the index wait queue
func (w *WordsIndexProcess) ReturnQueue() chan *WordsIndexRequest {
	return w.queue
}

// Add a words index reuest to queue
func (w *WordsIndexProcess) Add(req *WordsIndexRequest) (err error) {
	if w.count == WORDS_PROCESS_QUEUE_CAP {
		err = fmt.Errorf("The queue is full.")
		return
	}
	w.queue <- req
	w.count++
	return
}

// go to index
func (w *WordsIndexProcess) goindex() {
	for {
		if w.closed == true {
			return
		}

		req := <-w.queue
		switch req.Type {
		case WORDS_INDEX_TYPE_PAGE:
			// if is the page
			w.indexPage(req)
		case WORDS_INDEX_TYPE_AROUND:
			// if is the around link
			w.indexAroundLink(req)
		default:
			continue
		}
	}
}

// the page index
func (w *WordsIndexProcess) indexPage(req *WordsIndexRequest) {
	// TODO operate the KeyWords
	// operate the BodyContent
	word_split := toSequence(req.PageData.BodyContent)
	page_sentences := &PageSentences{
		Url:       req.PageData.Url,
		Sentences: word_split,
	}
	page_sentences.New(req.PageData.Url)
	/*
		exist, err := w.sentencedb.TableExist(req.PageData.Url)
		if err != nil {
			fmt.Println(err)
			return
		}
		if exist == false {
			err = w.sentencedb.NewTable(req.PageData.Url, &PageData{}, "Ver")
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		ver, err := w.sentencedb.Insert(req.PageData.Url, req.PageData)
		if err != nil {
			fmt.Println(err)
			return
		}*/
	exist, errd := w.sentencedb.drule.ExistRole(w.sentencedb.area, req.PageData.Url)
	if errd.IsError() != nil {
		fmt.Println(errd.IsError())
		return
	}
	if exist == true {
		var old_split map[uint64][]string
		errd = w.sentencedb.drule.ReadData(w.sentencedb.area, req.PageData.Url, "Sentences", &old_split)
		if errd.IsError() != nil {
			fmt.Println(errd.IsError())
			return
		}
		w.todelindex(req.PageData.Domain, req.PageData.Url, old_split)
	}
	errd = w.sentencedb.drule.StoreRole(w.sentencedb.area, page_sentences)
	if errd.IsError() != nil {
		fmt.Println(errd.IsError())
		return
	}
	w.toindex(req.PageData.Domain, req.PageData.Url, word_split)
}

func (w *WordsIndexProcess) todelindex(domain, url string, split map[uint64][]string) {
	for _, sentence := range split {
		for i, word := range sentence {
			// the sentence len
			slen := len(sentence)
			// this is one word
			tran, errd := w.worddb.drule.Begin()
			if errd.IsError() != nil {
				fmt.Println(errd.IsError())
				break
			}
			// check if the word already exist
			exist, errd := tran.ExistRole(w.worddb.area, word)
			if errd.IsError() != nil {
				tran.Rollback()
				fmt.Println(errd.IsError())
				break
			}
			if exist == false {
				tran.Rollback()
				break
			}
			// add the next word index
			if i < slen-1 {
				next_word := sentence[i+1]
				// check the context if exist
				cexist, errd := tran.ExistContext(w.worddb.area, word, next_word)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				if cexist == true {
					// read already exist index set, is status 0 string
					var index_set string
					have, errd := tran.ReadContextStatus(w.worddb.area, word, next_word, roles.CONTEXT_DOWN, url, 0, &index_set)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
					if have == true {
						errd = tran.DeleteContextBind(w.worddb.area, word, next_word, roles.CONTEXT_DOWN, url)
						if errd.IsError() != nil {
							tran.Rollback()
							fmt.Println(errd.IsError())
							break
						}
					}
				}
			}

			// add the prev word index
			if i > 1 {
				prev_word := sentence[i-1]
				// check the context if exist
				cexist, errd := tran.ExistContext(w.worddb.area, word, prev_word)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				if cexist == true {
					// read already exist index set, is status 0 string
					var index_set string
					have, errd := tran.ReadContextStatus(w.worddb.area, word, prev_word, roles.CONTEXT_UP, url, 0, &index_set)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
					if have == true {
						errd = tran.DeleteContextBind(w.worddb.area, word, prev_word, roles.CONTEXT_UP, url)
						if errd.IsError() != nil {
							tran.Rollback()
							fmt.Println(errd.IsError())
							break
						}
					}
				}
			}

			tran.Commit()
		}
	}
}

// to do the word index
func (w *WordsIndexProcess) toindex(domain, url string, split map[uint64][]string) {
	tmproles := make(map[string]*WordIndex)
	for count, sentence := range split {
		// the sentence len
		slen := len(sentence)
		// this is one sentence
		for i, word := range sentence {
			// check if have the word
			_, exist := tmproles[word]
			if exist == false {
				theindex := &WordIndex{}
				theindex.New(word)
				tmproles[word] = theindex
			}

			// add the next word index
			if i < slen-1 {
				next_word := sentence[i+1]
				exist := tmproles[word].ExistContext(next_word)
				if exist == false {
					err := tmproles[word].NewContext(next_word)
					if err != nil {
						fmt.Println(err)
						break
					}
				}
				// the word's count
				word_count := int64(count) + int64(i+1)
				// read already exist index set, is status 0 string
				var index_set string
				have, err := tmproles[word].GetContextStatus(next_word, roles.CONTEXT_DOWN, url, 0, &index_set)
				if err != nil {
					fmt.Println(err)
					break
				}
				if have == false {
					index_set = strconv.FormatInt(word_count, 10)
				} else {
					index_set = index_set + " " + strconv.FormatInt(word_count, 10)
				}
				err = tmproles[word].SetContextStatus(next_word, roles.CONTEXT_DOWN, url, 0, index_set)
				if err != nil {
					fmt.Println(err)
					break
				}
			}
			// add the prev word index
			if i > 1 {
				prev_word := sentence[i-1]
				exist := tmproles[word].ExistContext(prev_word)
				if exist == false {
					err := tmproles[word].NewContext(prev_word)
					if err != nil {
						fmt.Println(err)
						break
					}
				}
				// the word's count
				word_count := int64(count) + int64(i-1)
				// read already exist index set, is status 0 string
				var index_set string
				have, err := tmproles[word].GetContextStatus(prev_word, roles.CONTEXT_UP, url, 0, &index_set)
				if err != nil {
					fmt.Println(err)
					break
				}
				if have == false {
					index_set = strconv.FormatInt(word_count, 10)
				} else {
					index_set = index_set + " " + strconv.FormatInt(word_count, 10)
				}
				err = tmproles[word].SetContextStatus(prev_word, roles.CONTEXT_UP, url, 0, index_set)
				if err != nil {
					fmt.Println(err)
					break
				}
			}
		}
	}

	for word, _ := range tmproles {
		// this is one word
		tran, errd := w.worddb.drule.Begin()
		if errd.IsError() != nil {
			fmt.Println(errd.IsError())
			break
		}

		// check if the word already exist
		exist, errd := tran.ExistRole(w.worddb.area, word)
		if errd.IsError() != nil {
			tran.Rollback()
			fmt.Println(errd.IsError())
			break
		}

		if exist == false {
			errd = tran.StoreRole(w.worddb.area, tmproles[word])
			if errd.IsError() != nil {
				tran.Rollback()
				fmt.Println(errd.IsError())
				break
			}
		} else {
			// lock the role
			errd = tran.LockRole(w.worddb.area, word)
			if errd.IsError() != nil {
				tran.Rollback()
				fmt.Println(errd.IsError())
				break
			}
			allcontext := tmproles[word].GetContextsName()
			for _, onecontext := range allcontext {
				// check the context if exist
				cexist, errd := tran.ExistContext(w.worddb.area, word, onecontext)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				contextbody, _ := tmproles[word].GetContext(onecontext)
				if cexist == false {
					errd = tran.WriteContext(w.worddb.area, word, onecontext, contextbody)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
				} else {
					oldbody, _, errd := tran.ReadContext(w.worddb.area, word, onecontext)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
					newdownone, have := contextbody.Down[url]
					if have == true {
						oldone, have2 := oldbody.Down[url]
						if have2 == true {
							oldbody.Down[url].String[0] = oldone.String[0] + " " + newdownone.String[0]
						}
					}
					newupone, have := contextbody.Up[url]
					if have == true {
						oldone, have2 := oldbody.Up[url]
						if have2 == true {
							oldbody.Up[url].String[0] = oldone.String[0] + " " + newupone.String[0]
						}
					}

					errd = tran.WriteContext(w.worddb.area, word, onecontext, oldbody)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
				}
			}
		}

		tran.Commit()
	}

	//	for count, sentence := range split {
	//		// the sentence len
	//		slen := len(sentence)
	//		// this is one sentence
	//		for i, word := range sentence {
	//			// this is one word
	//			tran, errd := w.worddb.drule.Begin()
	//			if errd.IsError() != nil {
	//				fmt.Println(errd.IsError())
	//				break
	//			}
	//			// check if the word already exist
	//			exist, errd := tran.ExistRole(w.worddb.area, word)
	//			if errd.IsError() != nil {
	//				tran.Rollback()
	//				fmt.Println(errd.IsError())
	//				break
	//			}
	//			if exist == false {
	//				// if not exist, create it
	//				wordindex := &WordIndex{}
	//				wordindex.New(word)
	//				errd = tran.StoreRole(w.worddb.area, wordindex)
	//				if errd.IsError() != nil {
	//					tran.Rollback()
	//					fmt.Println(errd.IsError())
	//					break
	//				}
	//			}
	//			// add the next word index
	//			if i < slen-1 {
	//				next_word := sentence[i+1]
	//				// check the context if exist
	//				cexist, errd := tran.ExistContext(w.worddb.area, word, next_word)
	//				if errd.IsError() != nil {
	//					tran.Rollback()
	//					fmt.Println(errd.IsError())
	//					break
	//				}
	//				if cexist == false {
	//					errd = tran.CreateContext(w.worddb.area, word, next_word)
	//					if errd.IsError() != nil {
	//						tran.Rollback()
	//						fmt.Println(errd.IsError())
	//						break
	//					}
	//				}
	//				// the word's count
	//				word_count := int64(count) + int64(i+1)
	//				// read already exist index set, is status 0 string
	//				var index_set string
	//				have, errd := tran.ReadContextStatus(w.worddb.area, word, next_word, roles.CONTEXT_DOWN, url, 0, &index_set)
	//				if errd.IsError() != nil {
	//					tran.Rollback()
	//					fmt.Println(errd.IsError())
	//					break
	//				}
	//				if have == false {
	//					// if not have
	//					index_set = strconv.FormatInt(word_count, 10)
	//				} else {
	//					index_set = index_set + " " + strconv.FormatInt(word_count, 10)
	//				}
	//				errd = tran.WriteContextStatus(w.worddb.area, word, next_word, roles.CONTEXT_DOWN, url, 0, index_set)
	//				if errd.IsError() != nil {
	//					tran.Rollback()
	//					fmt.Println(errd.IsError())
	//					break
	//				}
	//			}

	//			// add the prev word index
	//			if i > 1 {
	//				prev_word := sentence[i-1]
	//				// check the context if exist
	//				cexist, errd := tran.ExistContext(w.worddb.area, word, prev_word)
	//				if errd.IsError() != nil {
	//					tran.Rollback()
	//					fmt.Println(errd.IsError())
	//					break
	//				}
	//				if cexist == false {
	//					errd = tran.CreateContext(w.worddb.area, word, prev_word)
	//					if errd.IsError() != nil {
	//						tran.Rollback()
	//						fmt.Println(errd.IsError())
	//						break
	//					}
	//				}
	//				// the word's count
	//				word_count := int64(count) + int64(i-1)
	//				// read already exist index set, is status 0 string
	//				var index_set string
	//				have, errd := tran.ReadContextStatus(w.worddb.area, word, prev_word, roles.CONTEXT_UP, url, 0, &index_set)
	//				if errd.IsError() != nil {
	//					tran.Rollback()
	//					fmt.Println(errd.IsError())
	//					break
	//				}
	//				if have == false {
	//					// if not have
	//					index_set = strconv.FormatInt(word_count, 10)
	//				} else {
	//					index_set = index_set + " " + strconv.FormatInt(word_count, 10)
	//				}
	//				errd = tran.WriteContextStatus(w.worddb.area, word, prev_word, roles.CONTEXT_UP, url, 0, index_set)
	//				if errd.IsError() != nil {
	//					tran.Rollback()
	//					fmt.Println(errd.IsError())
	//					break
	//				}
	//			}

	//			tran.Commit()
	//		}
	//	}
}

// the around link index
func (w *WordsIndexProcess) indexAroundLink(req *WordsIndexRequest) {
	// TODO
}
