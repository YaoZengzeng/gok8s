package processlistener

import (
	"fmt"
	"sync"
)

type addNotification struct {
	newObj interface{}
}

type updateNotification struct {
	oldObj, newObj interface{}
}

type deleteNotification struct {
	oldObj interface{}
}

type ResourceEventHandler interface {
	OnAdd(newObj interface{})
	OnUpdate(newObj interface{}, oldObj interface{})
	OnDelete(oldObj interface{})
}

type ResourceEventHandlerFuncs struct {
	AddFunc func(newObj interface{})
	UpdateFunc func(newObj interface{}, oldObj interface{})
	DeleteFunc func(oldObj interface{})
}

func (r *ResourceEventHandlerFuncs) OnAdd(newObj interface{}) {
	if r.AddFunc != nil {
		r.AddFunc(newObj)
	}
}

func (r *ResourceEventHandlerFuncs) OnUpdate(newObj interface{}, oldObj interface{}) {
	if r.UpdateFunc != nil {
		r.UpdateFunc(newObj, oldObj)
	}
}

func (r *ResourceEventHandlerFuncs) OnDelete(oldObj interface{}) {
	if r.DeleteFunc != nil {
		r.DeleteFunc(oldObj)
	}	
}

type ProcessListener struct {
	nextCh	chan interface{}
	addCh	chan interface{}
	handler ResourceEventHandler
	pendingNotifications	[]interface{}
}

func NewProcessListener(handler ResourceEventHandler) *ProcessListener {
	return &ProcessListener {
		nextCh:	make(chan interface{}),
		addCh:	make(chan interface{}),
		handler:	handler,
		pendingNotifications:	[]interface{}{},
	}
}

func (p *ProcessListener) add(notification interface{}) {
	p.addCh <- notification
}

func (p *ProcessListener) pop() {
	var notification interface{}
	var notifyCh chan interface{}
	defer close(p.nextCh)

	for {
		select {
		case notifyCh <- notification:
			if len(p.pendingNotifications) != 0 {
				notification = p.pendingNotifications[0]
				p.pendingNotifications = p.pendingNotifications[1:]
			} else {
				notifyCh = nil
			}
		case notificationToAdd, ok := <- p.addCh:
			if !ok {
				return
			}
			if notifyCh != nil {
				p.pendingNotifications = append(p.pendingNotifications, notificationToAdd)
			} else {
				notification = notificationToAdd
				notifyCh = p.nextCh
			}
		}
	}
}

func (p *ProcessListener) run() {
	for n := range p.nextCh {
		switch notification := n.(type) {
		case addNotification:
			p.handler.OnAdd(notification.newObj)
		case updateNotification:
			p.handler.OnUpdate(notification.newObj, notification.oldObj)
		case deleteNotification:
			p.handler.OnDelete(notification.oldObj)
		default:
			fmt.Println("pop unknown type")
		}
	}
}

type WaitGroup struct {
	wg sync.WaitGroup
}

func (wg *WaitGroup) Start(f func()) {
	wg.wg.Add(1)
	go func() {
		f()
		wg.wg.Done()
	}()
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}
