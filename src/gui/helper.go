package gui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// GetPoolList returns the list of pools available to the GUI miner
func (gui *GUI) GetPoolList() ([]PoolData, error) {
	var pools []PoolData
	resp, err := http.Get(fmt.Sprintf("%s/pool-list", gui.config.APIEndpoint))
	if err != nil {
		return pools, err
	}
	err = json.NewDecoder(resp.Body).Decode(&pools)
	if err != nil {
		return pools, err
	}

	return pools, nil
}

// GetPool returns a single pool's information
func (gui *GUI) GetPool(id int) (PoolData, error) {
	var pool PoolData
	resp, err := http.Get(fmt.Sprintf("%s/pool/%d", gui.config.APIEndpoint, id))
	if err != nil {
		return pool, err
	}
	err = json.NewDecoder(resp.Body).Decode(&pool)
	if err != nil {
		return pool, err
	}
	return pool, nil
}

// SaveConfig saves the configuration to disk
// TODO: Specify path here
func (gui *GUI) SaveConfig(config Config) error {
	configBytes, err := json.Marshal(&config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("config.json", configBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// GetStats returns stats for the interface. It requires the miner's
// hashrate to calculate XTL per dat
func (gui *GUI) GetStats(
	poolID int,
	hashrate float64,
	mid string) (string, error) {

	if mid == "" || poolID == 0 {
		return "", errors.New("No data yet")
	}
	resp, err := http.Get(
		fmt.Sprintf("%s/stats?pool=%d&hr=%.2f&mid=%s",
			gui.config.APIEndpoint,
			poolID,
			hashrate,
			mid))
	if err != nil {
		return "", err
	}
	statBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var stats GlobalStats
	err = json.Unmarshal(statBytes, &stats)
	if err != nil {
		return "", err
	}

	poolTemplate, err := gui.GetPoolTemplate(true)
	if err != nil {
		log.Fatalf("Unable to load pool template: '%s'", err)
	}
	poolData := PoolData{
		ID:        stats.Pool.ID,
		Hashrate:  stats.Pool.Hashrate,
		LastBlock: stats.Pool.LastBlock,
		Miners:    stats.Pool.Miners,
		URL:       stats.Pool.URL,
		Name:      stats.Pool.Name,
	}
	var templateHTML bytes.Buffer
	err = poolTemplate.Execute(&templateHTML, poolData)
	if err != nil {
		log.Fatalf("Unable to load pool template: '%s'", err)
	}
	stats.PoolHTML = templateHTML.String()

	statBytes, err = json.Marshal(&stats)
	if err != nil {
		return "", err
	}
	return string(statBytes), nil
}
