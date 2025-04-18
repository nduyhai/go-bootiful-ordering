table "orders" {
  schema = schema.public
  column "id" {
    type = varchar(36)
    null = false
  }
  column "customer_id" {
    type = varchar(36)
    null = false
  }
  column "status" {
    type = integer
    null = false
  }
  column "total_amount" {
    type = bigint
    null = false
  }
  column "created_at" {
    type = timestamp
    null = false
  }
  column "updated_at" {
    type = timestamp
    null = false
  }
  primary_key {
    columns = [column.id]
  }
}

table "order_items" {
  schema = schema.public
  column "id" {
    type = bigserial
    null = false
  }
  column "order_id" {
    type = varchar(36)
    null = false
  }
  column "product_id" {
    type = varchar(36)
    null = false
  }
  column "quantity" {
    type = integer
    null = false
  }
  column "price" {
    type = bigint
    null = false
  }
  column "created_at" {
    type = timestamp
    null = false
  }
  column "updated_at" {
    type = timestamp
    null = false
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key {
    columns = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete = CASCADE
  }
  index "idx_order_items_order_id" {
    columns = [column.order_id]
  }
}

schema "public" {}