<template>
    <div class="rooms-container app-border-r">
        <h2>Chats</h2>
        <slot name="rooms-header"></slot>
        <div class="box-search">
            <div class="icon-search" v-if="chats.length">
                <svg-icon name="search"/>
            </div>
            <input
                    type="search"
                    :placeholder="chats.SEARCH"
                    autocomplete="off"
                    @input="searchChats"
                    v-show="chats.length"
            />
            <div v-if="showAddChat" class="svg-button add-icon" @click="addChat">
                <svg-icon name="add"/>
            </div>
        </div>

        <loader :show="loadingChats"></loader>

        <div v-if="!loadingChats" class="room-list">
            <app-chat-miniature v-for="chat of chats"
                                :key="chat.chatId"
                                :chat="chat"
                                @click.native="openChat(chat)">
            </app-chat-miniature>
        </div>
    </div>
</template>

<script>
    import gql from 'graphql-tag';
    import ChatMiniature from "@/components/ChatMiniature";
    import SvgIcon from "@/components/SvgIcon";
    import Loader from "@/components/Loader";

    export default {
        name: "ChatList",
        components: {
            'app-chat-miniature': ChatMiniature,
            'svg-icon': SvgIcon,
            'loader': Loader,
        },
        data() {
            return {
                chats: [],
            };
        },
        props: {
            selectedChatId: {type: String},
            showAddChat: {type: Boolean, default: true},
            showChatList: {type: Boolean, default: true},
            loadingChats: {type: Boolean, default: false},
        },
        methods: {
            openChat(chat) {
                this.$emit('selectDiffrentChat', chat.chatId)
            },
            searchChats() {
                return null
            },
            addChat() {
              return null
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

<style lang="scss" scoped>
    .rooms-container {
        flex: 0 0 25%;
        min-width: 260px;
        max-width: 500px;
        position: relative;
        background: var(--chat-sidemenu-bg-color);
        height: 100%;
        overflow-y: auto;
        border-top-left-radius: var(--chat-container-border-radius);
        border-bottom-left-radius: var(--chat-container-border-radius);
    }

    .rooms-container-full {
        flex: 0 0 100%;
        max-width: 100%;
    }

    .box-search {
        display: flex;
        align-items: center;
        height: 64px;
        padding: 0 20px;
        margin-top: 5px;
    }

    .icon-search {
        display: flex;
        position: absolute;
        left: 30px;
        margin-top: 1px;

        svg {
            width: 22px;
            height: 22px;
        }
    }

    .add-icon {
        margin-left: auto;
        padding-left: 20px;
    }

    input {
        background: var(--chat-bg-color-input);
        color: var(--chat-color);
        border-radius: 4px;
        width: 100%;
        font-size: 15px;
        outline: 0;
        caret-color: var(--chat-color-caret);
        padding: 10px 10px 10px 38px;
        border: 1px solid var(--chat-sidemenu-border-color-search);
        border-radius: 20px;

        &::placeholder {
            color: var(--chat-color-placeholder);
        }
    }

    .rooms-empty {
        font-size: 14px;
        color: #9ca6af;
        font-style: italic;
        text-align: center;
        margin: 40px 0;
        line-height: 20px;
        white-space: pre-line;
    }

    .room-list {
        flex: 1;
        position: relative;
        max-width: 100%;
        cursor: pointer;
        padding: 5px 10px;
    }

    .room-item {
        border-radius: 8px;
        align-items: center;
        display: flex;
        flex: 1 1 100%;
        margin-bottom: 5px;
        padding: 0 16px;
        position: relative;
        min-height: 71px;

        &:hover {
            background: var(--chat-sidemenu-bg-color-hover);
            transition: background-color 0.3s cubic-bezier(0.25, 0.8, 0.5, 1);
        }

        &:not(:hover) {
            transition: background-color 0.3s cubic-bezier(0.25, 0.8, 0.5, 1);
        }
    }

    .room-selected {
        color: var(--chat-sidemenu-color-active) !important;
        background: var(--chat-sidemenu-bg-color-active) !important;

        &:hover {
            background: var(--chat-sidemenu-bg-color-active) !important;
        }
    }

    .name-container {
        flex: 1;
    }

    .title-container {
        display: flex;
        align-items: center;
        line-height: 25px;
    }

    .text-ellipsis {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .room-name {
        flex: 1;
        color: var(--chat-room-color-username);
        font-weight: 500;
    }

    .text-last {
        color: var(--chat-room-color-message);
        font-size: 12px;
    }

    .message-new {
        color: var(--chat-room-color-username);
        font-weight: 500;
    }

    .text-date {
        margin-left: 5px;
        font-size: 11px;
        color: var(--chat-room-color-timestamp);
    }

    .icon-check {
        height: 14px;
        width: 14px;
        vertical-align: middle;
        margin-top: -2px;
        margin-right: 1px;
    }

    .state-circle {
        width: 9px;
        height: 9px;
        border-radius: 50%;
        background-color: var(--chat-room-color-offline);
        margin-right: 6px;
        transition: 0.3s;
    }

    .state-online {
        background-color: var(--chat-room-color-online);
    }

    @media only screen and (max-width: 768px) {
        .box-search {
            height: 50px;
        }
        input {
            padding: 8px 8px 8px 38px;
        }
        .room-item {
            min-height: 60px;
            padding: 0 8px;
        }
    }
</style>
