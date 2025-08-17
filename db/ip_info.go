package db

import (
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/cortze/ragno/models"
)

// UpsertIP attemtps to insert IP in the DB - or Updates the data info if they where already there
func (p *PostgresDBService) UpsertIpInfo(IPInfo models.IPInfo) (query string, args []interface{}) {
	query = `
	INSERT INTO ip_info(
		ip,
		expiration_time,
		continent,
		continent_code,
		country,
		country_code,
		region,
		region_name,
		city,
		zip,
		lat,
		lon,
		isp,
		org,
		as_raw,
		asname,
		mobile,
		proxy,
		hosting)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
	ON CONFLICT (ip)
	DO UPDATE SET
		expiration_time = excluded.expiration_time,
		continent = excluded.continent,
		continent_code = excluded.continent_code,
		country = excluded.country,
		country_code = excluded.country_code,
		region = excluded.region,
		region_name = excluded.region_name,
		city = excluded.city,
		zip = excluded.zip,
		lat = excluded.lat,
		lon = excluded.lon,
		isp = excluded.isp,
		org = excluded.org,
		as_raw = excluded.as_raw,
		asname = excluded.asname,
		mobile = excluded.mobile,
		proxy = excluded.proxy,
		hosting = excluded.hosting;
	`

	args = append(args, IPInfo.IP)
	args = append(args, IPInfo.ExpirationTime)
	args = append(args, IPInfo.Continent)
	args = append(args, IPInfo.ContinentCode)
	args = append(args, IPInfo.Country)
	args = append(args, IPInfo.CountryCode)
	args = append(args, IPInfo.Region)
	args = append(args, IPInfo.RegionName)
	args = append(args, IPInfo.City)
	args = append(args, IPInfo.Zip)
	args = append(args, IPInfo.Lat)
	args = append(args, IPInfo.Lon)
	args = append(args, IPInfo.Isp)
	args = append(args, IPInfo.Org)
	args = append(args, IPInfo.As)
	args = append(args, IPInfo.AsName)
	args = append(args, IPInfo.Mobile)
	args = append(args, IPInfo.Proxy)
	args = append(args, IPInfo.Hosting)

	return query, args
}

// GetIPInfo reads all the information available for that specific IP in the DB
func (p *PostgresDBService) GetIPInfo(ip string) (models.IPInfo, error) {
	var ipInfo models.IPInfo
	err := p.psqlPool.QueryRow(p.ctx, `
		SELECT
			ip,
			expiration_time,
			continent,
			continent_code,
			country,
			country_code,
			region,
			region_name,
			city,
			zip,
			lat,
			lon,
			isp,
			org,
			as_raw,
			asname,
			mobile,
			proxy,
			hosting
		FROM ip_info
		WHERE ip=$1
	`, ip).Scan(
		&ipInfo.IP,
		&ipInfo.ExpirationTime,
		&ipInfo.Continent,
		&ipInfo.ContinentCode,
		&ipInfo.Country,
		&ipInfo.CountryCode,
		&ipInfo.Region,
		&ipInfo.RegionName,
		&ipInfo.City,
		&ipInfo.Zip,
		&ipInfo.Lat,
		&ipInfo.Lon,
		&ipInfo.Isp,
		&ipInfo.Org,
		&ipInfo.As,
		&ipInfo.AsName,
		&ipInfo.Mobile,
		&ipInfo.Proxy,
		&ipInfo.Hosting,
	)
	if err != nil {
		return models.IPInfo{}, err
	}

	return ipInfo, nil

}

// GetExpiredIpInfo returns all the IP whos' TTL has already expired
func (p *PostgresDBService) GetExpiredIPInfo() ([]string, error) {
	expIp_info := make([]string, 0)
	ipRows, err := p.psqlPool.Query(p.ctx, `
		SELECT ip
		FROM ip_info
		WHERE expiration_time < NOW();
	`)
	if err != nil {
		return expIp_info, errors.Wrap(err, "unable to get expired ip records")
	}

	defer ipRows.Close()

	for ipRows.Next() {
		var ip string
		err := ipRows.Scan(&ip)
		if err != nil {
			return expIp_info, errors.Wrap(err, "error parsing readed row for expired ip records")
		}
		expIp_info = append(expIp_info, ip)
	}

	return expIp_info, nil
}

// CheckIPRecords checks if a given IP is already stored in the DB as whether its TTL has expired
func (p *PostgresDBService) CheckIPRecords(ip string) (exists bool, expired bool, err error) {
	var readIp string
	var expTime time.Time

	row := p.psqlPool.QueryRow(p.ctx, `
		SELECT
			ip,
			expiration_time
		FROM ip_info
		WHERE ip=$1;
	`, ip)

	err = row.Scan(&readIp, &expTime)
	if err == pgx.ErrNoRows {
		return false, false, nil
	} else if err != nil {
		return
	}
	exists = (readIp == ip)
	expired = (expTime.Before(time.Now()))

	return
}

func (p *PostgresDBService) PersistIPInfo(ip models.IPInfo) {
	pAttempt := NewPersistable()
	pAttempt.query, pAttempt.values = p.UpsertIpInfo(ip)
	p.writeChan <- pAttempt
}
