<template>
    <form class="col-12"
          v-on:submit.prevent="onPostClick">

        <div class="input-group">
            <input type="text"
                   class="form-control"
                   placeholder="Message..."
                   v-model.trim="messageInput">
            <div class="input-group-append">
                <button class="btn btn-outline-secondary"
                        type="submit">Post</button>
            </div>
        </div>

    </form>
</template>

<script>
    import gql from 'graphql-tag';

    export default {
        name: "MessageForm",
        data() {
            return {
                messageInput: '',
            };
        },
        props: {
            selectedChatId: {type: String}
        },
        methods: {
            onPostClick() {
                const messageInput = this.messageInput;
                this.$apollo
                    .mutate({
                        mutation: gql`mutation($chatID: String!, $text: String!) {postMessage(chatID: $chatID, text: $text) {textMessage}}`,
                        variables: {
                            chatID: this.props.selectedChatId,
                            text: messageInput,
                        },
                    })
                    .then(() => {
                        this.messageInput = '';
                    })
                    .catch((e) => {
                        console.error(e);
                    });
            },
        },
    };
</script>

<style scoped>

</style>
