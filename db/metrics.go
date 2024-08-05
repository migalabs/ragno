package db

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	LastActivityValidRange = 180
) // 6 months

func (db *PostgresDBService) GetClientDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	cliDist := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
			SELECT
				client_name, count(client_name) as count
			FROM node_info
			WHERE
				first_connected IS NOT NULL AND
				deprecated = 'false' AND
				client_name IS NOT NULL AND
				last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
			GROUP BY client_name
			ORDER BY count DESC;
		`,
		LastActivityValidRange,
	)

	// close rows AND free the connection/session
	defer rows.Close()
	if err != nil {
		fmt.Print("\n", err.Error())
		return cliDist, errors.Wrap(err, "Unable to fetch client distribution")
	}

	for rows.Next() {
		var cliName string
		var count int
		err = rows.Scan(&cliName, &count)
		if err != nil {
			return cliDist, errors.Wrap(err, "unable to parse fetch client distribution")
		}
		cliDist[cliName] = count
	}

	return cliDist, nil
}

// Basic call over the whole list of non-deprecated peers
func (db *PostgresDBService) GetVersionDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	verDist := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
			SELECT client_name,
				client_raw_version,
				count(client_raw_version) as cnt
			FROM node_info
			WHERE
				first_connected IS NOT NULL AND
				deprecated = 'false' AND
				client_name IS NOT NULL AND
				last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
			GROUP BY client_name, client_raw_version
			ORDER BY client_name DESC, cnt DESC;
		`,
		LastActivityValidRange,
	)
	// make sure we close the rows AND we free the connection/session
	defer rows.Close()
	if err != nil {
		return verDist, errors.Wrap(err, "unable to fetch client distribution")
	}

	for rows.Next() {
		var cliName string
		var cliVersion string
		var count int
		err = rows.Scan(&cliName, &cliVersion, &count)
		if err != nil {
			return verDist, errors.Wrap(err, "unable to parse fetch client distribution")
		}
		verDist[cliName+"_"+cliVersion] = count
	}

	return verDist, nil
}

// Basic call over the whole list of non-deprecated peers
func (db *PostgresDBService) GetGeoDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	geoDist := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
			SELECT
				aux.country_code as country_code,
				count(aux.country_code) as cnt
			FROM (
				SELECT node_info.node_id, ip_info.ip, ip_info.country_code
				FROM node_info
				RIGHT JOIN ip_info on node_info.ip = ip_info.ip
				WHERE first_connected IS NOT NULL AND
					deprecated = 'false' AND
					client_name IS NOT NULL AND
					last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
			) as aux
			GROUP BY country_code
			ORDER BY cnt DESC;
		`,
		LastActivityValidRange,
	)
	// make sure we close the rows and we free the connection/session
	defer rows.Close()
	if err != nil {
		return geoDist, errors.Wrap(err, "unable to fetch client distribution")
	}

	for rows.Next() {
		var country string
		var count int
		err = rows.Scan(&country, &count)
		if err != nil {
			return geoDist, errors.Wrap(err, "unable to parse fetch client distribution")
		}
		geoDist[country] = count
	}

	return geoDist, nil
}

func (db *PostgresDBService) GetOsDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{}, 0)
	rows, err := db.psqlPool.Query(
		db.ctx,
		`
			SELECT
				client_os,
				count(client_os) as nodes
			FROM node_info
			WHERE first_connected IS NOT NULL AND
				deprecated = 'false' AND
				client_name IS NOT NULL AND
				last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
			GROUP BY client_os
			ORDER BY nodes DESC;
		`,
		LastActivityValidRange,
	)
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var os string
		var count int
		err = rows.Scan(&os, &count)
		if err != nil {
			return summary, err
		}
		summary[os] = count
	}
	return summary, nil
}

func (db *PostgresDBService) GetArchDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{}, 0)
	rows, err := db.psqlPool.Query(
		db.ctx,
		`
			SELECT
				client_arch,
				count(client_arch) as nodes
			FROM node_info
			WHERE first_connected IS NOT NULL AND
				deprecated = 'false' AND
				client_name IS NOT NULL AND
				last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
			GROUP BY client_arch
			ORDER BY nodes DESC;
		`,
		LastActivityValidRange,
	)
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var arch string
		var count int
		err = rows.Scan(&arch, &count)
		if err != nil {
			return summary, err
		}
		summary[arch] = count
	}
	return summary, nil
}

func (db *PostgresDBService) GetHostingDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{})
	// get the number of mobile hosts
	var mobile int
	err := db.psqlPool.QueryRow(
		db.ctx,
		`
		SELECT
			count(aux.mobile) as mobile
		FROM (
			SELECT
				ni.node_id,
				ni.first_connected,
				ni.client_name,
				ni.deprecated,
				ni.ip,
				ip_info.mobile
			FROM node_info as ni
			INNER JOIN ip_info ON ni.ip=ip_info.ip
			WHERE ni.deprecated='false' and
			      first_connected IS NOT NULL AND
			      client_name IS NOT NULL and
			      ip_info.mobile='true' and
			      last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as aux
		`,
		LastActivityValidRange,
	).Scan(&mobile)
	if err != nil {
		return summary, err
	}
	summary["mobile_ip_info"] = mobile

	// get the number of proxy peers
	var proxy int
	err = db.psqlPool.QueryRow(
		db.ctx,
		`
		SELECT
			count(aux.proxy) as under_proxy
		FROM (
			SELECT
				ni.node_id,
				ni.first_connected,
				ni.client_name,
				ni.deprecated,
				ni.ip,
				ip_info.proxy
			FROM node_info as ni
			INNER JOIN ip_info ON ni.ip=ip_info.ip
			WHERE ni.deprecated='false' and
			      first_connected IS NOT NULL AND
			      client_name IS NOT NULL and ip_info.proxy='true' and
			      last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as aux
		`,
		LastActivityValidRange,
	).Scan(&proxy)
	if err != nil {
		return summary, err
	}
	summary["under_proxy"] = proxy

	// get the number of hosted IP_info
	var hosted int
	err = db.psqlPool.QueryRow(
		db.ctx,
		`
		SELECT
			count(aux.hosting) as hosted_ip_info
		FROM (
			SELECT
				ni.node_id,
				ni.first_connected,
				ni.client_name,
				ni.deprecated,
				ni.ip,
				ip_info.hosting
			FROM node_info as ni
			INNER JOIN ip_info ON ni.ip=ip_info.ip
			WHERE ni.deprecated='false' and
			      first_connected IS NOT NULL AND
			      client_name IS NOT NULL and
			      ip_info.hosting='true' and
			      last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as aux
		`,
		LastActivityValidRange,
	).Scan(&hosted)
	if err != nil {
		return summary, err
	}
	summary["hosted_ip_info"] = hosted
	return summary, nil
}

func (db *PostgresDBService) GetIPDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
			SELECT
				nodes as nodes_per_ip,
				count(t.nodes) as number_of_ip_info
			FROM (
				SELECT
					ip,
					count(ip) as nodes
				FROM node_info
				WHERE deprecated = false AND
					client_name IS NOT NULL AND
					last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
				GROUP BY ip
				ORDER BY nodes DESC
			) as t
			GROUP BY nodes
			ORDER BY number_of_ip_info DESC;
		`,
		LastActivityValidRange,
	)
	if err != nil {
		return summary, err
	}

	for rows.Next() {
		var nodesPerIP int
		var ip_info int
		err = rows.Scan(
			&nodesPerIP,
			&ip_info,
		)
		if err != nil {
			return summary, err
		}
		summary[fmt.Sprintf("%d", nodesPerIP)] = ip_info
	}
	return summary, nil
}

func (db *PostgresDBService) GetRTTDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
			SELECT
				t.latency as latency_range,
				count(*) as nodes
			FROM (
				SELECT
					CASE
						WHEN latency between 0 AND 100 THEN ' 0-100ms'
						WHEN latency between 101 AND 200 THEN '101-200ms'
						WHEN latency between 201 AND 300 THEN '201-300ms'
						WHEN latency between 301 AND 400 THEN '301-400ms'
						WHEN latency between 401 AND 500 THEN '401-500ms'
						WHEN latency between 501 AND 600 THEN '501-600ms'
						WHEN latency between 601 AND 700 THEN '601-700ms'
						WHEN latency between 701 AND 800 THEN '701-800ms'
						WHEN latency between 801 AND 900 THEN '801-900ms'
						WHEN latency between 901 AND 1000 THEN '901-1000ms'
						ELSE '+1s'
					END as latency
				FROM node_info
				WHERE deprecated = false AND
					client_name IS NOT NULL AND
					last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
			) as t
			GROUP BY t.latency
			ORDER BY nodes DESC;
		`,
		LastActivityValidRange,
	)
	if err != nil {
		return summary, err
	}

	for rows.Next() {
		var rttRange string
		var rttValue int
		err = rows.Scan(
			&rttRange,
			&rttValue,
		)
		if err != nil {
			return summary, err
		}
		summary[rttRange] = rttValue
	}
	return summary, nil
}

func (db *PostgresDBService) GetDeprecatedNodes() (int, error) {
	log.Debug("fetching deprecated node count")

	var deprecatedCount int
	err := db.psqlPool.QueryRow(
		db.ctx,
		`
			SELECT
				count(deprecated)
			FROM node_info
			WHERE deprecated = 'true';
		`).Scan(
		&deprecatedCount,
	)
	if err != nil {
		return deprecatedCount, errors.Wrap(err, "unable to fetch deprecated node count")
	}

	return deprecatedCount, nil
}
