table "products" {
  schema = schema.public
  column "id" {
    type = varchar(36)
    null = false
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "description" {
    type = text
    null = true
  }
  column "price" {
    type = bigint
    null = false
  }
  column "stock" {
    type = integer
    null = false
  }
  column "category" {
    type = varchar(100)
    null = true
  }
  column "status" {
    type = integer
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

schema "public" {}