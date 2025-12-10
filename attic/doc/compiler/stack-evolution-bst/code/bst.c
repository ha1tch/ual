#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>

// Node structure for Binary Search Tree
typedef struct Node {
    int key;                // Key for searching
    void* value;            // Generic value pointer
    struct Node* left;      // Pointer to left child
    struct Node* right;     // Pointer to right child
    struct Node* parent;    // Pointer to parent (for easier traversal)
} Node;

// Binary Search Tree structure
typedef struct BST {
    Node* root;             // Pointer to root node
    int size;               // Number of nodes in the tree
} BST;

// Create a new node
Node* createNode(int key, void* value) {
    Node* newNode = (Node*)malloc(sizeof(Node));
    if (newNode == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(1);
    }
    
    newNode->key = key;
    newNode->value = value;
    newNode->left = NULL;
    newNode->right = NULL;
    newNode->parent = NULL;
    
    return newNode;
}

// Create a new Binary Search Tree
BST* createBST() {
    BST* tree = (BST*)malloc(sizeof(BST));
    if (tree == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(1);
    }
    
    tree->root = NULL;
    tree->size = 0;
    
    return tree;
}

// Insert a key-value pair into the tree
void insert(BST* tree, int key, void* value) {
    // Create new node
    Node* newNode = createNode(key, value);
    
    // If tree is empty, make new node the root
    if (tree->root == NULL) {
        tree->root = newNode;
        tree->size = 1;
        return;
    }
    
    // Find the appropriate position for the new node
    Node* current = tree->root;
    Node* parent = NULL;
    
    while (current != NULL) {
        parent = current;
        
        // If key already exists, update value and free the new node
        if (key == current->key) {
            current->value = value;
            free(newNode);
            return;
        }
        
        // Go left or right based on key comparison
        if (key < current->key) {
            current = current->left;
        } else {
            current = current->right;
        }
    }
    
    // Insert new node
    newNode->parent = parent;
    
    if (key < parent->key) {
        parent->left = newNode;
    } else {
        parent->right = newNode;
    }
    
    tree->size++;
}

// Find a node by key
Node* findNode(BST* tree, int key) {
    Node* current = tree->root;
    
    while (current != NULL) {
        if (key == current->key) {
            return current;
        }
        
        if (key < current->key) {
            current = current->left;
        } else {
            current = current->right;
        }
    }
    
    return NULL;  // Node not found
}

// Find a value by key
void* find(BST* tree, int key) {
    Node* node = findNode(tree, key);
    return (node != NULL) ? node->value : NULL;
}

// Find the minimum node in a subtree
Node* findMin(Node* node) {
    if (node == NULL) {
        return NULL;
    }
    
    while (node->left != NULL) {
        node = node->left;
    }
    
    return node;
}

// Find the maximum node in a subtree
Node* findMax(Node* node) {
    if (node == NULL) {
        return NULL;
    }
    
    while (node->right != NULL) {
        node = node->right;
    }
    
    return node;
}

// Get the successor of a node (next node in in-order traversal)
Node* successor(Node* node) {
    if (node == NULL) {
        return NULL;
    }
    
    // If right subtree exists, successor is the minimum node in right subtree
    if (node->right != NULL) {
        return findMin(node->right);
    }
    
    // Otherwise, find the nearest ancestor where node is in its left subtree
    Node* parent = node->parent;
    while (parent != NULL && node == parent->right) {
        node = parent;
        parent = parent->parent;
    }
    
    return parent;
}

// Delete a node with the given key
bool delete(BST* tree, int key) {
    // Find the node to delete
    Node* nodeToDelete = findNode(tree, key);
    
    // If node not found, return false
    if (nodeToDelete == NULL) {
        return false;
    }
    
    // Case 1: Node has no children (leaf node)
    if (nodeToDelete->left == NULL && nodeToDelete->right == NULL) {
        if (nodeToDelete->parent == NULL) {
            // Node is root
            tree->root = NULL;
        } else if (nodeToDelete == nodeToDelete->parent->left) {
            nodeToDelete->parent->left = NULL;
        } else {
            nodeToDelete->parent->right = NULL;
        }
        
        free(nodeToDelete);
    }
    // Case 2: Node has one child
    else if (nodeToDelete->left == NULL) {
        // Has right child only
        if (nodeToDelete->parent == NULL) {
            // Node is root
            tree->root = nodeToDelete->right;
            nodeToDelete->right->parent = NULL;
        } else if (nodeToDelete == nodeToDelete->parent->left) {
            nodeToDelete->parent->left = nodeToDelete->right;
            nodeToDelete->right->parent = nodeToDelete->parent;
        } else {
            nodeToDelete->parent->right = nodeToDelete->right;
            nodeToDelete->right->parent = nodeToDelete->parent;
        }
        
        free(nodeToDelete);
    }
    else if (nodeToDelete->right == NULL) {
        // Has left child only
        if (nodeToDelete->parent == NULL) {
            // Node is root
            tree->root = nodeToDelete->left;
            nodeToDelete->left->parent = NULL;
        } else if (nodeToDelete == nodeToDelete->parent->left) {
            nodeToDelete->parent->left = nodeToDelete->left;
            nodeToDelete->left->parent = nodeToDelete->parent;
        } else {
            nodeToDelete->parent->right = nodeToDelete->left;
            nodeToDelete->left->parent = nodeToDelete->parent;
        }
        
        free(nodeToDelete);
    }
    // Case 3: Node has two children
    else {
        // Find successor (minimum node in right subtree)
        Node* successor = findMin(nodeToDelete->right);
        
        // Copy successor's data to node being deleted
        nodeToDelete->key = successor->key;
        nodeToDelete->value = successor->value;
        
        // Delete successor (which has at most one child)
        if (successor->parent == nodeToDelete) {
            // Successor is direct right child
            nodeToDelete->right = successor->right;
            if (successor->right != NULL) {
                successor->right->parent = nodeToDelete;
            }
        } else {
            // Successor is deeper in right subtree
            successor->parent->left = successor->right;
            if (successor->right != NULL) {
                successor->right->parent = successor->parent;
            }
        }
        
        free(successor);
    }
    
    tree->size--;
    return true;
}

// Inorder traversal helper function
void inorderTraversalHelper(Node* node, void (*callback)(int key, void* value)) {
    if (node == NULL) {
        return;
    }
    
    inorderTraversalHelper(node->left, callback);
    callback(node->key, node->value);
    inorderTraversalHelper(node->right, callback);
}

// Inorder traversal of the tree
void inorderTraversal(BST* tree, void (*callback)(int key, void* value)) {
    inorderTraversalHelper(tree->root, callback);
}

// Level order traversal (BFS) of the tree
void levelOrderTraversal(BST* tree, void (*callback)(int key, void* value)) {
    if (tree->root == NULL) {
        return;
    }
    
    // Create a queue for BFS
    Node** queue = (Node**)malloc(tree->size * sizeof(Node*));
    if (queue == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(1);
    }
    
    int front = 0, rear = 0;
    queue[rear++] = tree->root;
    
    while (front < rear) {
        // Dequeue a node
        Node* current = queue[front++];
        
        // Process the current node
        callback(current->key, current->value);
        
        // Enqueue left child
        if (current->left != NULL) {
            queue[rear++] = current->left;
        }
        
        // Enqueue right child
        if (current->right != NULL) {
            queue[rear++] = current->right;
        }
    }
    
    free(queue);
}

// Calculate height of a node
int nodeHeight(Node* node) {
    if (node == NULL) {
        return 0;
    }
    
    int leftHeight = nodeHeight(node->left);
    int rightHeight = nodeHeight(node->right);
    
    return (leftHeight > rightHeight ? leftHeight : rightHeight) + 1;
}

// Calculate height of the tree
int height(BST* tree) {
    return nodeHeight(tree->root);
}

// Print a node with indentation for tree structure
void printNodeIndented(Node* node, int level, char* prefix) {
    if (node == NULL) {
        return;
    }
    
    // Print right subtree first (so it appears at the top)
    printNodeIndented(node->right, level + 1, "R:");
    
    // Print current node
    for (int i = 0; i < level; i++) {
        printf("    ");
    }
    printf("%s %d\n", prefix, node->key);
    
    // Print left subtree
    printNodeIndented(node->left, level + 1, "L:");
}

// Print the tree structure
void printTree(BST* tree) {
    printf("Binary Search Tree (size: %d)\n", tree->size);
    
    if (tree->root == NULL) {
        printf("  (empty)\n");
        return;
    }
    
    printNodeIndented(tree->root, 0, "Root:");
}

// Free memory for a subtree
void freeSubtree(Node* node) {
    if (node == NULL) {
        return;
    }
    
    // Post-order traversal to delete nodes
    freeSubtree(node->left);
    freeSubtree(node->right);
    free(node);
}

// Clear the tree (remove all nodes)
void clearTree(BST* tree) {
    freeSubtree(tree->root);
    tree->root = NULL;
    tree->size = 0;
}

// Free memory for the tree
void destroyTree(BST* tree) {
    clearTree(tree);
    free(tree);
}

// Get the size of the tree
int size(BST* tree) {
    return tree->size;
}

// Check if the tree is empty
bool isEmpty(BST* tree) {
    return tree->root == NULL;
}

// Example of how to use a BST for string values
// Note: This example assumes strings are dynamically allocated and
// should be freed when no longer needed
void printKeyValue(int key, void* value) {
    printf("Key: %d, Value: %s\n", key, (char*)value);
}

// Iterator structure for in-order traversal
typedef struct BSTIterator {
    Node** stack;       // Stack to keep track of nodes
    int top;            // Top of stack index
    int capacity;       // Maximum capacity of stack
    Node* current;      // Current node for iteration
} BSTIterator;

// Create a new iterator
BSTIterator* createIterator(BST* tree) {
    BSTIterator* iterator = (BSTIterator*)malloc(sizeof(BSTIterator));
    if (iterator == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(1);
    }
    
    // Allocate stack with size equal to tree height (maximum needed)
    int treeHeight = height(tree);
    iterator->stack = (Node**)malloc(treeHeight * sizeof(Node*));
    if (iterator->stack == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        free(iterator);
        exit(1);
    }
    
    iterator->top = -1;
    iterator->capacity = treeHeight;
    iterator->current = tree->root;
    
    return iterator;
}

// Check if iterator has more elements
bool hasNext(BSTIterator* iterator) {
    return (iterator->top >= 0 || iterator->current != NULL);
}

// Get the next key-value pair
bool next(BSTIterator* iterator, int* key, void** value) {
    if (!hasNext(iterator)) {
        return false;
    }
    
    // Reach leftmost node from current
    while (iterator->current != NULL) {
        // Push to stack
        iterator->stack[++iterator->top] = iterator->current;
        iterator->current = iterator->current->left;
    }
    
    // Get top node from stack
    Node* node = iterator->stack[iterator->top--];
    
    // Set result
    *key = node->key;
    *value = node->value;
    
    // Move to right subtree for next iteration
    iterator->current = node->right;
    
    return true;
}

// Free iterator
void destroyIterator(BSTIterator* iterator) {
    free(iterator->stack);
    free(iterator);
}

// Example usage
int main() {
    // Create a BST
    BST* tree = createBST();
    
    // Insert some key-value pairs
    insert(tree, 50, strdup("Fifty"));
    insert(tree, 30, strdup("Thirty"));
    insert(tree, 70, strdup("Seventy"));
    insert(tree, 20, strdup("Twenty"));
    insert(tree, 40, strdup("Forty"));
    insert(tree, 60, strdup("Sixty"));
    insert(tree, 80, strdup("Eighty"));
    
    // Print the tree
    printf("Original Tree:\n");
    printTree(tree);
    printf("\n");
    
    // Find a value
    printf("Finding key 40: %s\n", (char*)find(tree, 40));
    printf("Finding key 90: %s\n", (char*)find(tree, 90) ? (char*)find(tree, 90) : "Not found");
    printf("\n");
    
    // In-order traversal
    printf("In-order traversal:\n");
    inorderTraversal(tree, printKeyValue);
    printf("\n");
    
    // Level-order traversal
    printf("Level-order traversal:\n");
    levelOrderTraversal(tree, printKeyValue);
    printf("\n");
    
    // Delete a node
    printf("Deleting key 30...\n");
    delete(tree, 30);
    
    printf("Tree after deletion:\n");
    printTree(tree);
    printf("\n");
    
    // Iterate through the tree
    printf("Iteration through tree:\n");
    BSTIterator* iterator = createIterator(tree);
    int key;
    void* value;
    while (next(iterator, &key, &value)) {
        printf("Key: %d, Value: %s\n", key, (char*)value);
    }
    destroyIterator(iterator);
    printf("\n");
    
    // Free values (since we used strdup)
    inorderTraversal(tree, (void (*)(int, void*))free);
    
    // Clean up
    destroyTree(tree);
    
    return 0;
}