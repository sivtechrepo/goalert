import http from 'k6/http'
import Chance from 'https://chancejs.com/chance.min.js'

// Instantiate Chance so it can be used
var gen = new Chance()

class IDFetchType {
  constructor(c, id, queryName) {
    this.c = c
    this.id = id
    this.queryName = queryName
  }

  simpleField(fieldName) {
    return this.c.query(
      `query($id: ID!){${this.queryName}(id: $id){${fieldName}}}`,
      { id: this.id },
    ).data[this.queryName][fieldName]
  }
  simpleFieldMap(fieldName, Type) {
    return this.c
      .query(`query($id: ID!){${this.queryName}(id: $id){${fieldName}{id}}}`, {
        id: this.id,
      })
      .data[this.queryName][fieldName].map((obj) => new Type(this.c, obj.id))
  }
  simpleUpdateField(fieldName, typeName, value) {
    const name = this.queryName[0].toUpperCase() + this.queryName.slice(1)
    return this.c.query(
      `mutation($id: ID!, $value: ${typeName}){
        update${name}(input:{id: $id, ${fieldName}: $value})
      }`,
      { id: this.id, value },
    )
  }

  delete() {
    return this.c.query(
      `mutation($id: ID!){
        deleteAll(input:[{id: $id, type: ${this.queryName}}])
      }`,
      { id: this.id },
    )
  }
}

class UserContactMethod extends IDFetchType {
  constructor(c, id) {
    super(c, id, 'userContactMethod')
  }

  get name() {
    return this.simpleField('name')
  }
  set name(newName) {
    this.simpleUpdateField('name', 'String!', newName)
  }

  get type() {
    return this.simpleField('type')
  }
  get value() {
    return this.simpleField('value')
  }
  get formattedValue() {
    return this.simpleField('formattedValue')
  }
  get disabled() {
    return this.simpleField('disabled')
  }
  get lastTestVerifyAt() {
    return this.simpleField('lastTestVerifyAt')
  }
}

class Service extends IDFetchType {
  constructor(c, id) {
    super(c, id, 'service')
  }
  get name() {
    return this.simpleField('name')
  }
  set name(newName) {
    this.simpleUpdateField('name', 'String', newName)
  }
}

class EP extends IDFetchType {
  constructor(c, id) {
    super(c, id, 'escalationPolicy')
  }

  get name() {
    return this.simpleField('name')
  }
  set name(newName) {
    this.simpleUpdateField('name', 'String', newName)
  }

  get description() {
    return this.simpleField('description')
  }
  set description(value) {
    this.simpleUpdateField('description', 'String', value)
  }
}

class Rotation extends IDFetchType {
  constructor(c, id) {
    super(c, id, 'rotation')
  }
  get name() {
    return this.simpleField('name')
  }
  get activeUserIndex() {
    return this.simpleField('activeUserIndex')
  }
  set activeUserIndex(idx) {
    this.simpleUpdateField('activeUserIndex', 'Int!', idx)
  }

  get users() {
    return this.userIDs.map((id) => new User(this.c, id))
  }
  get userIDs() {
    return this.simpleField('userIDs')
  }
  set userIDs(ids) {
    this.simpleUpdateField('userIDs', '[ID!]', ids)
  }

  get timeZone() {
    return this.simpleField('timeZone')
  }
  set timeZone(name) {
    this.simpleUpdateField('timeZone', 'String', name)
  }
}

class User extends IDFetchType {
  constructor(c, id) {
    super(c, id, 'user')
  }

  get name() {
    return this.simpleField('name')
  }
  get email() {
    return this.simpleField('email')
  }
  get role() {
    return this.simpleField('role')
  }
  get statusUpdateContactMethodID() {
    return this.simpleField('statusUpdateContactMethodID')
  }
  get isFavorite() {
    return this.simpleField('isFavorite')
  }
  get contactMethods() {
    return this.simpleFieldMap('contactMethods', UserContactMethod)
  }
}

export class Client {
  constructor(baseURL) {
    this.baseURL = baseURL
    this.login()
  }

  login(user = 'admin', pass = 'admin123') {
    let resp = http.get(this.baseURL + '/api/v2/identity/providers')
    let providers = JSON.parse(resp.body)
    let loginURL = providers.find((p) => p.ID === 'basic').URL

    http.post(
      this.baseURL + loginURL,
      {
        username: user,
        password: pass,
      },
      {
        headers: {
          referer: this.baseURL,
        },
      },
    )
  }

  logout() {
    http.get(this.baseURL + '/api/v2/identity/logout')
  }

  service(id) {
    return new Service(this, id)
  }
  services() {
    return this.query(`query{services{nodes{id}}}`).data.services.nodes.map(
      (u) => this.service(u.id),
    )
  }
  randService() {
    return gen.pickone(this.services())
  }

  escalationPolicy(id) {
    return new EP(this, id)
  }
  escalationPolicies() {
    return this.query(
      `query{escalationPolicies{nodes{id}}}`,
    ).data.escalationPolicies.nodes.map((u) => this.escalationPolicy(u.id))
  }
  randEP() {
    return gen.pickone(this.escalationPolicies())
  }

  newEP() {
    const q = this.query(
      `mutation($input: CreateEscalationPolicyInput!){createEscalationPolicy(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          description: gen.sentence(),
          repeat: gen.integer({ min: 1, max: 5 }),
        },
      },
    )
    const id = q.data.createEscalationPolicy.id
    return new EP(this, id)
  }

  newUser() {
    const q = this.query(
      `mutation($input: CreateUserInput!){createUser(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.name(),
          email: gen.email(),
          role: gen.pickone(['user', 'admin']),
          username: gen.string({ alpha: true, length: 20, casing: 'lower' }),
          password: gen.string({ alpha: true, length: 20 }),
        },
      },
    )
    const id = q.data.createUser.id
    return new User(this, id)
  }

  newService(epID) {
    if (!epID) {
      epID = this.randEP().id
    }

    const q = this.query(
      `mutation($input: CreateServiceInput!){createService(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          description: gen.sentence(),
          escalationPolicyID: epID,
        },
      },
    )
    const id = q.data.createService.id
    return new Service(this, id)
  }

  newRotation() {
    let tz = gen.timezone()
    if (!tz.utc) {
      tz = 'Etc/UTC'
    } else {
      tz = tz.utc[0]
    }
    const q = this.query(
      `mutation($input: CreateRotationInput!){createRotation(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          description: gen.sentence(),
          timeZone: tz,
          start: gen.date().toISOString(),
          type: gen.pickone(['hourly', 'weekly']),
          shiftLength: gen.integer({ min: 1, max: 20 }),
        },
      },
    )
    const id = q.data.createRotation.id
    return new Rotation(this, id)
  }

  user(id) {
    if (!id) {
      id = this.query(`query{user{id}}`).data.user.id
    }
    return new User(this, id)
  }
  rotation(id) {
    return new Rotation(this, id)
  }

  randUser() {
    // don't return current user
    const users = this.users()
    const thisUser = this.user()
    while (true) {
      const u = gen.pickone(users)
      if (u.id === thisUser.id) {
        continue
      }

      return new User(this, u.id)
    }
  }
  users() {
    return this.query(`query{users{nodes{id}}}`).data.users.nodes.map((u) =>
      this.user(u.id),
    )
  }

  randRotation() {
    return gen.pickone(this.rotations())
  }
  rotations() {
    return this.query(`query{rotations{nodes{id}}}`).data.rotations.nodes.map(
      (u) => this.rotation(u.id),
    )
  }

  query(query, variables = {}) {
    const resp = http.post(
      this.baseURL + '/api/graphql',
      JSON.stringify({
        query,
        variables,
      }),
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )

    const res = JSON.parse(resp.body)

    if (res.errors) {
      throw new Error(res.errors[0].message)
    }

    return res
  }
}

export function newClient(url) {}
