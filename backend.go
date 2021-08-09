package main

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	etcdItemPrefix = "/items/"
)

type BackendStore interface {
	Get(key string) (string, bool, error)
	Set(key string, value string) error
	Delete(key string) error
}

type EtcdBackend struct {
	etcdClient *clientv3.Client
}

func NewEtcdBackend(endpoints []string) (EtcdBackend, error) {
	backend := EtcdBackend{}

	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://0.0.0.0:2379"},
	})
	if err != nil {
		return backend, err
	}
	backend.etcdClient = client
	return backend, nil
}

func (e *EtcdBackend) Get(key string) (string, bool, error) {
	etcdKey := etcdItemPrefix + key

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := e.etcdClient.Get(ctx, etcdKey)
	if err != nil {
		return "", false, err
	}

	if len(resp.Kvs) == 0 {
		// key is not set
		return "", false, nil
	}

	return string(resp.Kvs[0].Value), true, nil
}

func (e *EtcdBackend) Set(key string, value string) error {
	etcdKey := etcdItemPrefix + key

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := e.etcdClient.Put(ctx, etcdKey, value)
	if err != nil {
		return err
	}

	return nil
}

func (e *EtcdBackend) Delete(key string) error {
	etcdKey := etcdItemPrefix + key

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := e.etcdClient.Delete(ctx, etcdKey)
	if err != nil {
		return err
	}

	return nil
}
