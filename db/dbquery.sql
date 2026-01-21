



-- name: CurrentScriptId :one
SELECT 
    id 
FROM mission_script
WHERE isUsing = 1
LIMIT 1;

-- name: CurrentScriptName :one
SELECT 
    name 
FROM mission_script
WHERE isUsing = 1
LIMIT 1;



-- name: CreateConatiner :exec
INSERT INTO cargo_info (
    id, 
    custom_id,
    status, 
    owner,
    metadata,
    updatedAt,
    custom_cargo_metadata_id,
    conveyor_configId
) VALUES (
    ?,?,?,?,?,?,?,?
);


-- name: CreateConatinerHistory :exec
INSERT INTO cargo_history (
    id,
    cargo_id,
    action,
    actor,
    description
) VALUES (
    ?,?,?,?,?
);


-- name: ConveyorConfigId :one
SELECT cc.id
FROM peripheral_name pn
JOIN conveyor_config cc ON pn.id = cc.name
WHERE pn.name = ?;

-- name: PeripheralName :one
SELECT * FROM peripheral_name WHERE name = ?;