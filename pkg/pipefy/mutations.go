package pipefy

const CreateCardMutation = `mutation($input: CreateCardInput!) {
  createCard(input: $input) {
    card {
      id
      title
    }
  }
}`

const UpdateCardFieldMutation = `mutation($input: UpdateCardFieldInput!) {
  updateCardField(input: $input) {
    card {
      id
    }
    success
  }
}`
