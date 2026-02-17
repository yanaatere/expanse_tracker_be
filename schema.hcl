table "users" {
  schema = public
  column "id" {
    null = false
    type = serial
  }
  column "username" {
    null = false
    type = varchar(50)
  }
  column "email" {
    null = false
    type = varchar(100)
  }
  column "created_at" {
    null = true
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null = false
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  unique "users_email_key" {
    columns = [column.email]
  }
  unique "users_username_key" {
    columns = [column.username]
  }
}

table "categories" {
  schema = public
  column "id" {
    null = false
    type = serial
  }
  column "name" {
    null = false
    type = varchar(255)
  }
  column "description" {
    null = true
    type = text
  }
  column "created_at" {
    null = true
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null = false
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}

table "transactions" {
  schema = public
  column "id" {
    null = false
    type = serial
  }
  column "user_id" {
    null = false
    type = integer
  }
  column "type" {
    null = false
    type = varchar(50)
  }
  column "amount" {
    null = false
    type = numeric(10, 2)
  }
  column "description" {
    null = true
    type = text
  }
  column "category_id" {
    null = true
    type = integer
  }
  column "transaction_date" {
    null = false
    type = date
    default = sql("CURRENT_DATE")
  }
  column "created_at" {
    null = true
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null = false
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "transactions_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  foreign_key "transactions_category_id_fkey" {
    columns = [column.category_id]
    ref_columns = [table.categories.column.id]
    on_delete = SET_NULL
  }
  check "transactions_type_check" {
    expr = "type IN ('income', 'expense')"
  }
}

table "balances" {
  schema = public
  column "id" {
    null = false
    type = serial
  }
  column "user_id" {
    null = false
    type = integer
  }
  column "total_balance" {
    null = false
    type = numeric(12, 2)
    default = 0
  }
  column "created_at" {
    null = true
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null = false
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  unique "balances_user_id_key" {
    columns = [column.user_id]
  }
  foreign_key "balances_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
}
