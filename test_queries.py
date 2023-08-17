#!/usr/bin/env python3

# INSERT INTO t_node_info (
# 	node_id,
# 	peer_id,
# 	first_connected,
# 	last_connected,
# 	last_tried,
# 	client_name,
# 	capabilities,
# 	software_info,
# 	error
# ) VALUES (
# 	f3d4165dd5e1902d4204516c76475f931079a5004df4250d9b2294f1f65b2537,
# 	0,
# 	2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042,
# 	2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042,
# 	2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042,
# 	blabla,blabla1,blabla2,blabla3)
# ON CONFLICT (node_id) DO UPDATE SET
# 	node_id = f3d4165dd5e1902d4204516c76475f931079a5004df4250d9b2294f1f65b2537,
# 	peer_id = 0,
# 	last_connected = 2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042,
# 	last_tried = 2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042,
# 	client_name = blabla,
# 	capabilities = blabla1,
# 	software_info = blabla2,
# 	error = blabla3
# WHERE '' IS NULL OR $9 = '';

import psycopg2
from psycopg2.extras import execute_values


def main():
    # Your variable values
    var1 = 'f3d4165dd5e1902d4204516c76475f931079a5004df4250d9b2294f1f65b2537'
    var2 = '0'
    var3 = '2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042'
    var4 = '2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042'
    var5 = '2023-07-27 15:06:04.190516 +0200 CEST m=+1963.235948042'
    var6 = 'client_name'
    var7 = ['cap1', 'cap2', 'cap3']
    var8 = 'soft_info'
    var9 = ''


    capabilities_array = "{" + ",".join(f'"{c}"' for c in var7) + "}"

    # Establish a connection to the database
    conn = psycopg2.connect(database="el_nodes", user="postgres",
                            password="mysecretpassword", host="localhost", port="5432")

    # Create a cursor
    cur = conn.cursor()

    # Execute the query with variables
    cur.execute("""
        INSERT INTO t_node_info (
            node_id,
            peer_id,
            first_connected,
            last_connected,
            last_tried,
            client_name,
            capabilities,
            software_info,
            error
        ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (node_id) DO UPDATE SET
            node_id = %s,
            peer_id = %s,
            last_connected = %s,
            last_tried = %s,
            client_name = %s,
            capabilities = %s,
            software_info = %s,
            error = %s
        WHERE t_node_info.error IS NULL OR t_node_info.error = '';
    """, (var1, var2, var3, var4, var5, var6, capabilities_array, var8, var9,
          var1, var2, var4, var5, var6, capabilities_array, var8, var9))

    # Commit the transaction and close the connection
    conn.commit()
    cur.close()
    conn.close()


if __name__ == '__main__':
    main()
