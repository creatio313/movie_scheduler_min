resource "sakura_kms" "database_key" {
  name        = "データベース暗号化キー"
  description = "データベースディスクを暗号化するためのKMSキー。"
  key_origin  = "generated"
}

resource "sakura_secret_manager" "database_secret" {
  name        = "データベース認証情報シークレット"
  description = "データベース認証情報を格納するためのSecret Managerシークレット。"
  kms_key_id  = sakura_kms.database_key.id
}

resource "sakura_secret_manager_secret" "database_secret_value" {
  name     = "database_secret_value"
  vault_id = sakura_secret_manager.database_secret.id
  value_wo = jsonencode({
    database_name = var.database_username
    host          = var.database_ip
    port          = var.database_port
    username      = var.database_username
    password      = var.database_password
  })
  value_wo_version = 1
}

resource "sakura_bridge" "bridge_for_vps" {
  name        = "VPS接続用ブリッジ"
  description = "VPSとデータベースを接続するためのブリッジ。"
}

resource "sakura_vswitch" "switch_for_database" {
  name        = "データベース接続用スイッチ"
  description = "VPSとデータベースを接続するためのスイッチ。"

  bridge_id = sakura_bridge.bridge_for_vps.id
  icon_id   = var.database_icon
  zone      = var.zone
}

resource "sakura_database" "movie_scheduler_database" {
  name        = "撮影計画支援電算処理システムDB"
  description = "撮影計画支援電算処理システム用のMariaDB。"

  backup = {
    days_of_week = ["mon"]
    time         = "04:00"
  }

  network_interface = {
    vswitch_id    = sakura_vswitch.switch_for_database.id
    ip_address    = var.database_ip
    netmask       = 24
    gateway       = var.database_gateway
    port          = var.database_port
    source_ranges = var.database_source_ranges
  }

  username            = var.database_username
  password_wo         = var.database_password
  password_wo_version = 1

  database_type    = "mariadb"
  database_version = "10.11"

  disk = {
    encryption_algorithm = "aes256_xts"
    kms_key_id           = sakura_kms.database_key.id
  }

  icon_id = var.database_icon

  monitoring_suite = {
    enabled = true
  }
  # 基盤が弱いため、キャッシュはOFFにする
  parameters = {
    event_scheduler              = "OFF"
    innodb_buffer_pool_size      = 134217728
    log_warnings                 = 2
    long_query_time              = 10
    max_allowed_packet           = 16777216
    max_connections              = 100
    query_alloc_block_size       = 8192
    query_cache_limit            = 1048576
    query_cache_min_res_unit     = 4096
    query_cache_size             = 536870912
    query_cache_type             = 0
    query_cache_wlock_invalidate = "OFF"
    query_prealloc_size          = 8192
    slow_query_log               = "ON"
    sort_buffer_size             = 2097152
    sql_mode                     = "STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION"
    tmpdir                       = "/tmp"
  }
  plan = "10g"
  zone = var.zone
}

# Outputs for the application
output "database_secret_vault_id" {
  description = "The Vault ID of the database secret for Secret Manager"
  value       = sakura_secret_manager.database_secret.id
}

output "database_secret_name" {
  description = "The name of the database secret"
  value       = sakura_secret_manager_secret.database_secret_value.name
}