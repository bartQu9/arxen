<template>
    <div>
        <h3>Chats</h3>
        <app-chat-miniature v-for="chat of chats"
                            :key="chat.chatId"
                            :chat="chat"
                            @click.native="openChat(chat)">
        </app-chat-miniature>
    </div>
</template>

<script>
    import gql from 'graphql-tag';
    import ChatMiniature from "@/components/ChatMiniature";

    export default {
        name: "ChatList",
        components: {
            'app-chat-miniature': ChatMiniature,
        },
        data() {
            return {
                chats: [],
            };
        },
        props: {
            selectedChatId: {type: String}
        },
        methods: {
          openChat(chat) {
              this.$emit('selectDiffrentChat', chat.chatId)
          }
        },
        apollo: {
            chats() {
                //const chat = this.$currentChats();
                return {
                    query: gql`{ chats { chatId clientsIPsList } }`,
                    subscribeToMore: {
                        //  subscription($user: String!) { chatCreated(chat: $chat) }
                        document: gql`subscription{ chatCreated { chatId clientsIPsList } }`,
                        variables: {},
                        updateQuery: (prev, {subscriptionData}) => {
                            if (!subscriptionData.data) {
                                return prev;
                            }
                            const chat = subscriptionData.data.chatCreated;
                            if (prev.users.find((u) => u.chatId === chat.chatID)) {
                                return prev;
                            }
                            return Object.assign({}, prev, {
                                users: [chat, ...prev.users],
                            });
                        },
                    },
                };
            },
        },
    }


</script>

<style scoped>

</style>
