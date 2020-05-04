import gql from "graphql-tag";
<template>
    <div class="row">
        <app-chat-list class="col-2"
                       @selectDiffrentChat="updateSelectedChat"
                       :selectedChatId="selectedChatId"
        />
        <div class="col-10">
            <div v-if="selectedChatId">
                <div class="row mb-3">
                    <app-message-form :selectedChatId="selectedChatId"/>
                </div>

                <div class="row">
                    <app-message-list class="col-12"
                                      :selectedChatId="selectedChatId"
                    />
                </div>
            </div>
            <div v-if="!selectedChatId">Select Chat</div>
        </div>
    </div>
</template>

<script>
    import ChatList from '@/components/ChatList.vue';
    import MessageList from '@/components/MessageList.vue';
    import MessageForm from '@/components/MessageForm.vue';

    export default {
        components: {
            'app-chat-list': ChatList,
            'app-message-list': MessageList,
            'app-message-form': MessageForm,
        },
        data() {
            return {
                chatList: [],
                selectedChatId: '',
            };
        },
        methods: {
            updateSelectedChat(chatID) {
                this.selectedChatId = chatID;
            }
        }
    };
</script>
