<template>
    <div>
        <div v-if="selectedChatId && selectedChatId.length && messages">
            <app-message v-for="message of messages"
                         :key="message"
                         :message="message">
            </app-message>
        </div>
        <div v-if="!selectedChatId">
            Empty
        </div>
        <div v-if="!messages.length && selectedChatId.length">No Messages</div>
    </div>
</template>

<script>
    import gql from 'graphql-tag';
    import Message from '@/components/Message';

    export default {
        name: "MessageList",
        components: {
            'app-message': Message,
        },
        props: {
            selectedChatId: {type: String}
        },
        data() {
            return {
                messages: [],
            };
        },
        apollo: {
            messages() {
                return {
                    query: gql`
                    query($chatID: String!) {
                            messages(chatID: $chatID) {
                                chatId
                                user
                                text
                                timeStamp
                            }
                    }
                    `,
                    variables() { return { chatID: this.selectedChatId } },
                    subscribeToMore: {
                        document: gql`
                        subscription($chatID: String!) {
                            messagePosted(chatID: $chatID) {
                                chatId
                                user
                                text
                                timeStamp
                            }
                        }`,
                        variables() {
                            return {chatID: this.selectedChatId}
                        },
                        updateQuery: (prev, {subscriptionData}) => {
                            if (!subscriptionData.data) {
                                return prev;
                            }
                            const message = subscriptionData.data.messagePosted;
                            if (prev.messages.find((m) => m === message)) {
                                return prev;
                            }
                            return Object.assign({}, prev, {
                                messages: [message, ...prev.messages],
                            });
                        },
                    },
                };
            },
        },
    };
</script>

<style scoped>

</style>
